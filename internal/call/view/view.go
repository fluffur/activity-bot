package view

func FormatWelcomeCallMessageSet() string {
	return "Новое сообщение для call установлено"
}

func FormatCallOnJoinEnabled() string {
	return "Теперь при инвайте новых участников будет вызываться call"
}

func FormatCallOnJoinDisabled() string {
	return "Теперь при инвайте новых участников не будет вызываться call"
}

func FormatWelcomeCallMessage(message string) string {
	if message == "" {
		return "Сообщение ещё не указано"
	}
	return "Сообщение: " + message
}
