package config

import "github.com/PaulSonOfLars/gotgbot/v2"

var BotCommands = []gotgbot.BotCommand{
	{Command: "stats", Description: "📊 Недельный отчёт"},
	{Command: "inactive", Description: "💤 Неактивные участники"},
	{Command: "norm", Description: "📈 Норма сообщений"},
	{Command: "rest", Description: "🛌 Управление рестом"},
	{Command: "role", Description: "🏷️ Роль участника"},
	{Command: "me", Description: "👁️ Информация о себе"},
	{Command: "you", Description: "🧑 Информация об участнике"},
	{Command: "roles", Description: "🗂️ Список ролей"},
	{Command: "admins", Description: "🛡️ Администраторы бота"},
	{Command: "is_admin", Description: "🛡️ Проверить статус администратора"},
	{Command: "update", Description: "🔁 Обновить данные чата"},
	{Command: "newbie", Description: "🌱 Порог новичка"},
	{Command: "all", Description: "📣 Созвать участников"},
	{Command: "call_enable", Description: "📣 Включить созыв при входе новичка"},
	{Command: "call_disable", Description: "📣 Отключить созыв при входе новичка"},
	{Command: "call_message", Description: "💬 Сообщение для созыва"},
	{Command: "week_start", Description: "📅 День начала недели"},
	{Command: "help", Description: "🆘 Помощь"},
}
