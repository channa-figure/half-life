package cmd

import (
	"fmt"
	"os"
	"sync"
	"time"

	"gopkg.in/yaml.v2"
)

const (
	configFilePath                       = "./config.yaml"
	slashingPeriodUptimeWarningThreshold = 99.80
	slashingPeriodUptimeErrorThreshold   = 98
	recentBlocksToCheck                  = 20
	notifyEvery                          = 20 // check runs every ~30 seconds, so will notify for continued errors and rollup stats every ~10 mins
	recentMissedBlocksNotifyThreshold    = 10
	sentryGRPCErrorNotifyThreshold       = 1 // will notify with error for any more than this number of consecutive grpc errors for a given sentry
	sentryOutOfSyncErrorNotifyThreshold  = 1 // will notify with error for any more than this number of consecutive out of sync errors for a given sentry
	sentryHaltErrorNotifyThreshold       = 1 // will notify with error for any more than this number of consecutive halt errors for a given sentry
)

type AlertLevel int8

const (
	alertLevelNone AlertLevel = iota
	alertLevelWarning
	alertLevelHigh
	alertLevelCritical
)

type AlertType int8

const (
	alertTypeJailed AlertType = iota
	alertTypeTombstoned
	alertTypeOutOfSync
	alertTypeBlockFetch
	alertTypeMissedRecentBlocks
	alertTypeGenericRPC
	alertTypeHalt

	// leave this at the end for iteration
	alertTypeEnd
)

type SentryAlertType int8

const (
	sentryAlertTypeNone SentryAlertType = iota
	sentryAlertTypeGRPCError
	sentryAlertTypeOutOfSyncError
	sentryAlertTypeHalt
)

type SentryStats struct {
	Name            string
	Version         string
	Height          int64
	SentryAlertType SentryAlertType
}

type ValidatorStats struct {
	Timestamp                   time.Time
	Height                      int64
	RecentMissedBlocks          int64
	LastSignedBlockHeight       int64
	RecentMissedBlockAlertLevel AlertLevel
	LastSignedBlockTimestamp    time.Time
	SlashingPeriodUptime        float64
	SentryStats                 []*SentryStats
	AlertLevel                  AlertLevel
	RPCError                    bool
}

type ValidatorAlertState struct {
	AlertTypeCounts              map[AlertType]int64
	SentryGRPCErrorCounts        map[string]int64
	SentryOutOfSyncErrorCounts   map[string]int64
	SentryHaltErrorCounts        map[string]int64
	SentryLatestHeight           map[string]int64
	RecentMissedBlocksCounter    int64
	RecentMissedBlocksCounterMax int64
	LatestBlockChecked           int64
	LatestBlockSigned            int64
}

type ValidatorAlertNotification struct {
	Alerts         []string
	ClearedAlerts  []string
	NotifyForClear bool
	AlertLevel     AlertLevel
}

type NotificationsConfig struct {
	Service string                `yaml:"service"`
	Discord *DiscordChannelConfig `yaml:"discord"`
	Twilio  *TwilioConfig         `yaml:"twilio"`
}

type HalfLifeConfig struct {
	Notifications *NotificationsConfig `yaml:"notifications"`
	Validators    []*ValidatorMonitor  `yaml:"validators"`
}

type DiscordWebhookConfig struct {
	ID    string `yaml:"id"`
	Token string `yaml:"token"`
}

type DiscordChannelConfig struct {
	Webhook      DiscordWebhookConfig `yaml:"webhook"`
	AlertUserIDs []string             `yaml:"alert-user-ids"`
	Username     string               `yaml:"username"`
}
type TwilioConfig struct {
	AccountSid string `yaml:"account-sid"`
	AuthToken  string `yaml:"auth-token"`
	To         string `yaml:"to"`
	From       string `yaml:"from"`
}

type Sentry struct {
	Name string `yaml:"name"`
	GRPC string `yaml:"grpc"`
}

type ValidatorMonitor struct {
	Name                   string    `yaml:"name"`
	RPC                    string    `yaml:"rpc"`
	Address                string    `yaml:"address"`
	ChainID                string    `yaml:"chain-id"`
	DiscordStatusMessageID *string   `yaml:"discord-status-message-id"`
	RPCRetries             *int      `yaml:"rpc-retries"`
	Sentries               *[]Sentry `yaml:"sentries"`
}

func saveConfig(configFile string, config *HalfLifeConfig, writeConfigMutex *sync.Mutex) {
	writeConfigMutex.Lock()
	defer writeConfigMutex.Unlock()

	yamlBytes, err := yaml.Marshal(config)
	if err != nil {
		fmt.Printf("Error during config yaml marshal %v\n", err)
	}

	err = os.WriteFile(configFile, yamlBytes, 0644)
	if err != nil {
		fmt.Printf("Error saving config yaml %v\n", err)
	}
}
