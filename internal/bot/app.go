package bot

import (
	"activity-bot/internal/adapter"
	"activity-bot/internal/admin"
	"activity-bot/internal/chat"
	"activity-bot/internal/config"
	"activity-bot/internal/db/postgres"
	db "activity-bot/internal/db/postgres/sqlc"
	"activity-bot/internal/logger"
	"activity-bot/internal/member"
	"activity-bot/internal/rest"
	"activity-bot/internal/user"
	"context"
	"fmt"
	"log"

	"github.com/celestix/gotgproto"
	"github.com/celestix/gotgproto/dispatcher"
	"github.com/celestix/gotgproto/sessionMaker"
	"github.com/cohesion-org/deepseek-go"
	"github.com/glebarez/sqlite"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
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

	MemberService *member.Service
	AdminService  *admin.Service
	UserService   *user.Service
	ChatService   *chat.Service
	RestService   *rest.Service
}

func NewApp(cfg config.Config) (*App, error) {
	ctx := context.Background()

	logger.Init(cfg.Debug)

	bot, err := gotgproto.NewClient(
		cfg.AppID,
		cfg.AppHash,
		gotgproto.ClientTypeBot(cfg.BotToken),
		&gotgproto.ClientOpts{
			Session: sessionMaker.SqlSession(sqlite.Open(cfg.SQLSessionPath)),
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

	adminsProvider := adapter.NewTelegramChatMembersProvider(bot)

	userService := user.NewService(userRepo)
	memberService := member.NewService(memberRepo, chatRepo, userRepo, adminsProvider)
	adminService := admin.NewService(adminRepo, cfg.BotOwnerID)
	chatService := chat.NewService(chatRepo)
	restService := rest.NewService(restRepo)

	return &App{
		Config:        cfg,
		Pool:          pool,
		Rdb:           rdb,
		Deepseek:      deepseek.NewClient(cfg.DeepseekAPIKey),
		Bot:           bot,
		dp:            dp,
		AsyncClient:   client,
		AsyncServer:   srv,
		MemberService: memberService,
		AdminService:  adminService,
		UserService:   userService,
		ChatService:   chatService,
		RestService:   restService,
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
