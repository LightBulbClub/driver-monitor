package analysis

import (
	"fmt"
	"math"
	"time"

	"github.com/LightBulbClub/driver-monitor/config"
	"github.com/LightBulbClub/driver-monitor/data"
)

// DriverCooldown 记录每个司机的最近一次告警时间，用于静默处理
var DriverCooldown = make(map[string]time.Time)

// StartAlertEngine 启动实时分析和告警协程
func StartAlertEngine() {
	go processData()
	go processAlerts() // 启动告警通知协程
}

// processData 从 DataChannel 读取数据并执行分析
func processData() {
	for handbandData := range data.DataChannel {
		checkHeartRate(handbandData)
		checkAcceleration(handbandData)
		// TODO: 添加更多复杂的分析模型，如疲劳指数、微睡眠检测等
	}
}

// checkHeartRate 检测心率是否在正常区间
func checkHeartRate(d data.HandbandData) {
	if d.HeartRate < config.HeartRateMin || d.HeartRate > config.HeartRateMax {
		message := fmt.Sprintf("心率异常: %d bpm，超出正常范围 [%d, %d]",
			d.HeartRate, config.HeartRateMin, config.HeartRateMax)

		triggerAlert(d.DriverID, "HeartRateAnomaly", message)
	}
}

// checkAcceleration 检测加速度剧烈变化 (简易合向量检测)
func checkAcceleration(d data.HandbandData) {
	// 计算加速度合向量的平方
	accelMagnitudeSquared := d.AccelX*d.AccelX + d.AccelY*d.AccelY + d.AccelZ*d.AccelZ
	accelMagnitude := math.Sqrt(accelMagnitudeSquared)

	if accelMagnitude > config.AccelThreshold {
		message := fmt.Sprintf("加速度剧烈变化: %.2f (阈值 %.2f)，可能发生剧烈动作或碰撞。",
			accelMagnitude, config.AccelThreshold)

		triggerAlert(d.DriverID, "SuddenMovement", message)
	}
}

// triggerAlert 发送告警到 AlertChannel，并处理静默期
func triggerAlert(driverID, alertType, message string) {
	lastAlertTime := DriverCooldown[driverID]

	// 检查是否在静默期内
	if time.Since(lastAlertTime) < config.AlertCooldown {
		// fmt.Printf("DEBUG: Driver %s 处于告警静默期\n", driverID)
		return
	}

	newAlert := data.Alert{
		DriverID:  driverID,
		Timestamp: time.Now(),
		Message:   message,
		Type:      alertType,
	}

	// 更新告警时间并发送
	DriverCooldown[driverID] = time.Now()
	data.AlertChannel <- newAlert
}

// processAlerts 负责从 AlertChannel 中取出告警并执行通知
func processAlerts() {
	for alert := range data.AlertChannel {
		// TODO: 实际的通知逻辑：发送邮件、短信、App 推送（例如通过 MQTT 或 WebHook）
		fmt.Printf("🚨 ----------------------------------------------------------------\n")
		fmt.Printf("🚨 CRITICAL ALERT [%s] for Driver %s at %s\n",
			alert.Type, alert.DriverID, alert.Timestamp.Format(time.RFC3339))
		fmt.Printf("🚨 Message: %s\n", alert.Message)
		fmt.Printf("🚨 ----------------------------------------------------------------\n")
	}
}
