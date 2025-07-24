module GoopBot

go 1.24.5

require (
	github.com/bwmarrin/discordgo v0.29.0
	github.com/joho/godotenv v1.5.1
	github.com/mattn/go-sqlite3 v1.14.28
	github.com/redis/go-redis/v9 v9.11.0
	gorm.io/gorm v1.30.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	golang.org/x/crypto v0.0.0-20210421170649-83a5a9bb288b // indirect
	golang.org/x/sys v0.21.0 // indirect
	golang.org/x/text v0.27.0 // indirect
)

require gorm.io/driver/sqlite v1.6.0

replace (
	GoopBot/internal/bot/bot => D:/MyGit/GoopBot/GoopBot/internal/bot/bot
	GoopBot/internal/features => D:/MyGit/GoopBot/GoopBot/internal/features
)
