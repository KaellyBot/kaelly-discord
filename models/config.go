package models

const (
	ConfigFileName = ".env"

	// Discord Bot Token
	Token = "TOKEN"

	// Shard ID. More on https://discord.com/developers/docs/topics/gateway#sharding
	ShardId = "SHARD_ID"

	// Total number of shards used to run the entire application
	ShardCount = "SHARD_COUNT"
)

var (
	DefaultConfigValues = map[string]interface{}{
		Token:      "",
		ShardId:    0,
		ShardCount: 1,
	}
)