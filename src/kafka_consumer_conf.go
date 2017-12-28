package main

/* 程序运行所需的配置文件 */
type KafkaConsumerConf struct {
	PartionGroup       string `json:PartionGroup`       // kafka Group
	BrokersAddr        string `json:BrokersAddr`        // kafka Host
	Topics             string `json:Topics`             // kafka topic
	MessageCount       uint64 `json:MessageCount`       // need read message count
	NumProcessRoutine  int    `json:NumProcessRoutine`  // 处理消息的协程数
	NumConsumerRoutine int    `json:NumConsumerRoutine` // consumer的协程数
}
