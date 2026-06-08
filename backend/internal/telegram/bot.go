func (b *Bot) handleStart(chatID int64) {
    msg := "Proxy Manager Bot\nCommands:\n/users - list users\n/add - create user"
    b.send(chatID, msg)
}

func (b *Bot) handleAddUser(chatID int64, args string) {
    // создаём пользователя, возвращаем ссылку
    user := b.db.CreateUser(...)
    link := b.core.GetClientLink(user.UUID, "vless")
    b.send(chatID, link)
}