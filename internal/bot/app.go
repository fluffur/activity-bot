package bot

import (
	"activity-bot/internal/adapter"
	"activity-bot/internal/admin"
	"activity-bot/internal/call"
	"activity-bot/internal/chat"
	"activity-bot/internal/config"
	"activity-bot/internal/db/postgres"
	db "activity-bot/internal/db/postgres/sqlc"
	"activity-bot/internal/logger"
	"activity-bot/internal/member"
	msg "activity-bot/internal/message"
	"activity-bot/internal/rest"
	"activity-bot/internal/session"
	"activity-bot/internal/stats"
	"activity-bot/internal/user"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gotd/contrib/middleware/floodwait"
	"github.com/gotd/contrib/middleware/ratelimit"

	"github.com/celestix/gotgproto"
	"github.com/celestix/gotgproto/dispatcher"
	"github.com/celestix/gotgproto/sessionMaker"
	"github.com/cohesion-org/deepseek-go"
	"github.com/glebarez/sqlite"
	"github.com/gotd/td/telegram"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"golang.org/x/time/rate"
)

type App struct {
	Config      config.Config
	Pool        *pgxpool.Pool
	Rdb         *redis.Client
	Deepseek    *deepseek.Client
	Bot         *gotgproto.Client
	dp          dispatcher.Dispatcher
	AsyncClient *asynq.Client
	AsyncServer *asynq.Server

	MemberService  *member.Service
	AdminService   *admin.Service
	UserService    *user.Service
	ChatService    *chat.Service
	RestService    *rest.Service
	StatsService   *stats.Service
	MessageService *msg.Service
	CallService    *call.Service
	SessionService *session.Service
}

func NewApp(cfg config.Config) (*App, error) {
	ctx := context.Background()

	logger.Init(cfg.Debug)
	waiter := floodwait.NewWaiter().WithCallback(func(ctx context.Context, wait floodwait.FloodWait) {
		fmt.Printf("Waiting for flood, dur: %d\n", wait.Duration)
	})

	ratelimiter := ratelimit.New(rate.Every(time.Millisecond*100), 30)
	bot, err := gotgproto.NewClient(
		cfg.AppID,
		cfg.AppHash,
		gotgproto.ClientTypeBot(cfg.BotToken),
		&gotgproto.ClientOpts{
			Session:     sessionMaker.SqlSession(sqlite.Open(cfg.SQLSessionPath)),
			Middlewares: []telegram.Middleware{waiter, ratelimiter},
			RunMiddleware: func(origRun func(ctx context.Context, f func(ctx context.Context) error) (err error), ctx context.Context, f func(ctx context.Context) (err error)) (err error) {
				return origRun(ctx, func(ctx context.Context) error {
					return waiter.Run(ctx, f)
				})
			},
		},
	)
	if err != nil {
		log.Fatalln("failed to start client:", err)
	}

	dp := bot.Dispatcher
	pool, err := pgxpool.New(ctx, cfg.DBDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres pool: %w", err)
	}

	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisADDR})
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: cfg.RedisADDR})
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: cfg.RedisADDR},
		asynq.Config{Concurrency: 10},
	)

	queries := db.New(pool)

	memberRepo := postgres.NewMemberRepository(queries)
	adminRepo := postgres.NewAdminRepository(queries)
	chatRepo := postgres.NewChatRepository(queries)
	userRepo := postgres.NewUserRepository(queries)
	restRepo := postgres.NewRestRepository(queries, pool)
	statsRepo := postgres.NewStatsRepository(queries)
	messageRepo := postgres.NewMessageRepository(queries)
	sessionRepo := postgres.NewSessionRepository(queries)

	adminsProvider := adapter.NewTelegramChatMembersProvider(bot)

	userService := user.NewService(userRepo)
	memberService := member.NewService(memberRepo, chatRepo, userRepo, adminsProvider)
	adminService := admin.NewService(adminRepo, cfg.BotOwnerID)
	chatService := chat.NewService(chatRepo)
	restService := rest.NewService(restRepo)
	statsService := stats.NewService(statsRepo)
	messageService := msg.NewService(messageRepo)
	callService := call.NewService(chatRepo, memberService, statsService)
	sessionService := session.NewService(sessionRepo)

	return &App{
		Config:         cfg,
		Pool:           pool,
		Rdb:            rdb,
		Deepseek:       deepseek.NewClient(cfg.DeepseekAPIKey),
		Bot:            bot,
		dp:             dp,
		AsyncClient:    client,
		AsyncServer:    srv,
		MemberService:  memberService,
		AdminService:   adminService,
		UserService:    userService,
		ChatService:    chatService,
		RestService:    restService,
		StatsService:   statsService,
		MessageService: messageService,
		CallService:    callService,
		SessionService: sessionService,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	a.RegisterHandlers()
	return a.Bot.Idle()
}

func (a *App) Close() {
	if a.Pool != nil {
		a.Pool.Close()
	}
	if a.Rdb != nil {
		_ = a.Rdb.Close()
	}
	if a.AsyncClient != nil {
		_ = a.AsyncClient.Close()
	}
}
