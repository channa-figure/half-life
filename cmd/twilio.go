package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type TwilioNotificationService struct {
	accountSid  string
	authToken   string
	to          string
	from        string
	endpointUrl string
}

func NewTwilioNotificationService(accountSid, authToken, to, from string) *TwilioNotificationService {
	return &TwilioNotificationService{
		accountSid:  accountSid,
		authToken:   authToken,
		endpointUrl: fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", accountSid),
		to:          to,
		from:        from,
	}
}

func getCurrentStatsSms(stats ValidatorStats, vm *ValidatorMonitor) string {
	var uptime string
	if stats.SlashingPeriodUptime == 0 {
		uptime = "N/A"
	} else {
		uptime = fmt.Sprintf("%.02f", stats.SlashingPeriodUptime)
	}

	title := fmt.Sprintf("%s (%s%% up)", vm.Name, uptime)

	var description string
	sentryString := ""

	if vm.Sentries != nil {
		for _, vmSentry := range *vm.Sentries {
			sentryFound := false
			for _, sentryStats := range stats.SentryStats {
				if vmSentry.Name == sentryStats.Name {
					var statusIcon string
					if sentryStats.SentryAlertType == sentryAlertTypeNone {
						statusIcon = iconGood
					} else {
						statusIcon = iconError
					}

					var height string
					if sentryStats.Height == 0 {
						height = "N/A"
					} else {
						height = fmt.Sprint(sentryStats.Height)
					}
					var version string
					if sentryStats.Version == "" {
						version = "N/A"
					} else {
						version = sentryStats.Version
					}

					sentryString += fmt.Sprintf("\n%s **%s** - Height **%s** - Version **%s**", statusIcon, sentryStats.Name, height, version)
					sentryFound = true
					break
				}
			}
			if !sentryFound {
				sentryString += fmt.Sprintf("\n%s **%s** - Height **N/A** - Version **N/A**", iconError, vmSentry.Name)
			}
		}
	}

	recentSignedBlocks := fmt.Sprintf("%s Latest Blocks Signed: **N/A**", iconWarning)

	var latestBlock string
	if stats.Timestamp.Before(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)) {
		latestBlock = fmt.Sprintf("%s Height **N/A**", iconError)
	} else {
		var rpcStatusIcon string
		if stats.RPCError {
			rpcStatusIcon = iconError
		} else {
			rpcStatusIcon = iconGood
			var recentSignedBlocksIcon string
			if stats.RecentMissedBlockAlertLevel >= alertLevelHigh {
				recentSignedBlocksIcon = iconError
			} else if stats.RecentMissedBlockAlertLevel == alertLevelWarning {
				recentSignedBlocksIcon = iconWarning
			} else {
				recentSignedBlocksIcon = iconGood
			}
			recentSignedBlocks = fmt.Sprintf("%s Latest Blocks Signed: **%d/%d**", recentSignedBlocksIcon, recentBlocksToCheck-stats.RecentMissedBlocks, recentBlocksToCheck)

		}
		latestBlock = fmt.Sprintf("%s Height **%s** - **%s**", rpcStatusIcon, fmt.Sprint(stats.Height), formattedTime(stats.Timestamp))
	}

	if stats.Height == stats.LastSignedBlockHeight {
		description = fmt.Sprintf("%s\n%s%s",
			latestBlock, recentSignedBlocks, sentryString)
	} else {
		var lastSignedBlock string
		if stats.LastSignedBlockHeight == -1 {
			lastSignedBlock = fmt.Sprintf("%s Last Signed **N/A**", iconError)
		} else {
			lastSignedBlock = fmt.Sprintf("%s Last Signed **%s** - **%s**", iconError, fmt.Sprint(stats.LastSignedBlockHeight), formattedTime(stats.LastSignedBlockTimestamp))
		}
		description = fmt.Sprintf("%s\n%s\n%s%s",
			latestBlock, lastSignedBlock, recentSignedBlocks, sentryString)
	}
	return fmt.Sprintf("%s \n %s", title, description)
}

func (service *TwilioNotificationService) SendSMS(msg string) error {
	msgData := url.Values{}
	msgData.Set("To", service.to)
	msgData.Set("From", service.from)
	msgData.Set("Body", msg)
	msgDataReader := *strings.NewReader(msgData.Encode())
	client := &http.Client{}
	req, _ := http.NewRequest("POST", service.endpointUrl, &msgDataReader)
	req.SetBasicAuth(service.accountSid, service.authToken)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, _ := client.Do(req)
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		var data map[string]interface{}
		decoder := json.NewDecoder(resp.Body)
		err := decoder.Decode(&data)
		if err == nil {
			fmt.Println(data["sid"])
		}
		return nil
	} else {
		return errors.New(resp.Status)
	}
}

// implements NotificationService interface
func (service *TwilioNotificationService) UpdateValidatorRealtimeStatus(
	configFile string,
	config *HalfLifeConfig,
	vm *ValidatorMonitor,
	stats ValidatorStats,
	writeConfigMutex *sync.Mutex,
) {
	statusMsg := getCurrentStatsSms(stats, vm)
	// err := service.SendSMS(statusMsg)
	// if err != nil {
	// 	fmt.Printf("Error Sending Twilio message: %v\n", err)
	// 	return
	// }
	fmt.Printf("Twilio TODO: implement configurable period to send sms status message: %s", statusMsg)
}

// implements NotificationService interface
func (service *TwilioNotificationService) SendValidatorAlertNotification(
	config *HalfLifeConfig,
	vm *ValidatorMonitor,
	stats ValidatorStats,
	alertNotification *ValidatorAlertNotification,
) {
	var title string
	if stats.SlashingPeriodUptime > 0 {
		title = fmt.Sprintf("%s (%.02f%% up)", vm.Name, stats.SlashingPeriodUptime)
	} else {
		title = fmt.Sprintf("%s (N/A%% up)", vm.Name)
	}

	if len(alertNotification.Alerts) > 0 {
		alertString := ""
		for _, alert := range alertNotification.Alerts {
			alertString += fmt.Sprintf("\n• %s", alert)
		}
		err := service.SendSMS(fmt.Sprintf("%s\n**Errors:**\n%s", title, strings.Trim(alertString, "\n")))
		if err != nil {
			fmt.Printf("Error Sending Twilio message: %v\n", err)
			return
		}
	}

	if len(alertNotification.ClearedAlerts) > 0 {
		clearedAlertsString := ""
		for _, alert := range alertNotification.ClearedAlerts {
			clearedAlertsString += fmt.Sprintf("\n• %s", alert)
		}
		msg := fmt.Sprintf("%s\n**Errors cleared:**\n%s", title, strings.Trim(clearedAlertsString, "\n"))
		err := service.SendSMS(msg)
		if err != nil {
			fmt.Printf("Error Sending Twilio message: %v\n", err)
			return
		}
	}
}
