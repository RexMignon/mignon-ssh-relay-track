package config

import (
	"encoding/hex"
	"encoding/json" // 导入 JSON 包
	"fmt"
	"mignon-ssh-port-forworder-dev/app/pkg/constant"
	sm4 "mignon-ssh-port-forworder-dev/app/pkg/encryption_algorithm"
	log "mignon-ssh-port-forworder-dev/app/pkg/logging"
	"mignon-ssh-port-forworder-dev/app/pkg/utils"
	"os"
)

type (
	IConfigInterFace interface {
		AddIConfigGroup(group *IConfigGroup)
		ModifyIConfigGroup(Id string, group *IConfigGroup)
		ModifyIConfigLinkGroup(ServerId string, LinkGroupId string, group *IConfigLinkGroup)
		RemoveIConfigLinkGroup(ServerId string, LinkGroupId string)
		RemoveIConfigGroup(Id string)
		AddIConfigLinkGroup(Id string, group IConfigLinkGroup)
		SetValue()
	}

	IConfig struct {
		Config []IConfigGroup `json:"config"`
		// default value true, this is the theme switch
		IsDark    bool `json:"is_dark"`
		IsEnglish bool `json:"is_english"`
	}

	IConfigGroup struct {
		Id         string             `json:"id"`
		Username   string             `json:"username"`
		Password   string             `json:"password"`
		ServerName string             `json:"server_name"`
		ServerHost string             `json:"server_host"`
		ServerPort int                `json:"server_port"`
		LinkGroup  []IConfigLinkGroup `json:"link_group"`
		IsOpen     bool               `json:"is_open"`
		Notes      string             `json:"notes"`
	}

	// IConfigLinkGroup 此结构体是用来标记需要转发/穿透的名称
	IConfigLinkGroup struct {
		Id string `json:"id"`
		// 为转发, 穿透的实例标记名称
		Name string `json:"name"`
		// 需要向本地转发, 或者是向服务器穿透的本机Host, 如0.0.0.0
		LocalHost string `json:"local_host"`
		// 转发前的Host即服务器的host, 默认为127.0.0.1即可, 或者是向服务器穿透的服务器host
		RemoteHost string `json:"remote_host"`
		// 远程端口
		RemotePort int `json:"remote_port"`
		// 本地端口
		LocalPort int `json:"local_port"`
		// 注释
		Notes string `json:"notes"`
		// 是否是穿透
		IsPenetrate bool `json:"is_penetrate"`
		IsOpen      bool `json:"is_open"`
	}
)

// AddIConfigGroup 添加服务器组
func (config *IConfig) AddIConfigGroup(group *IConfigGroup) {
	config.Config = append(config.Config, *group)
	config.SetValue()
}

// ModifyIConfigGroup 修改指定的服务器组
func (config *IConfig) ModifyIConfigGroup(Id string, group *IConfigGroup) {
	index := -1
	for i, item := range config.Config {
		if item.Id == Id {
			index = i
			break
		}
	}
	if index == -1 {
		return
	}
	config.Config[index] = *group
	config.SetValue()
}

func (config *IConfig) ModifyIConfigLinkGroup(ServerId string, LinkGroupId string, group *IConfigLinkGroup) {
	serverIndex := -1
	for i, item := range config.Config {
		if item.Id == ServerId {
			serverIndex = i
			break
		}
	}
	if serverIndex == -1 {
		return
	}
	linkIndex := -1
	for i, item := range config.Config[serverIndex].LinkGroup {
		if item.Id == LinkGroupId {
			linkIndex = i
			break
		}
	}
	if linkIndex == -1 {
		return
	}
	config.Config[serverIndex].LinkGroup[linkIndex] = *group
	config.SetValue()
}

// RemoveIConfigLinkGroup 删除指定的
func (config *IConfig) RemoveIConfigLinkGroup(ServerId string, LinkGroupId string) {
	serverIndex := -1
	for i, item := range config.Config {
		if item.Id == ServerId {
			serverIndex = i
			break
		}
	}
	if serverIndex == -1 {
		return
	}
	linkIndex := -1
	for i, item := range config.Config[serverIndex].LinkGroup {
		if item.Id == LinkGroupId {
			linkIndex = i
			break
		}
	}
	if linkIndex == -1 {
		return
	}
	config.Config[serverIndex].LinkGroup = append(
		config.Config[serverIndex].LinkGroup[:linkIndex],
		config.Config[serverIndex].LinkGroup[linkIndex+1:]...,
	)
	config.SetValue()
}

// RemoveIConfigGroup 删除指定的服务器组
func (config *IConfig) RemoveIConfigGroup(Id string) {
	index := -1
	for i, item := range config.Config {
		if item.Id == Id {
			index = i
			break
		}
	}
	if index == -1 {
		return
	}
	config.Config = append(config.Config[:index], config.Config[index+1:]...)
	config.SetValue()
}

// AddIConfigLinkGroup 向Id为x的服务器组内添加IConfigLinkGroup
func (config *IConfig) AddIConfigLinkGroup(Id string, group IConfigLinkGroup) {
	index := -1
	for i, configGroup := range config.Config {
		if configGroup.Id == Id {
			index = i
			break
		}
	}
	if index == -1 {
		return
	}
	config.Config[index].LinkGroup = append(config.Config[index].LinkGroup, group)
	config.SetValue()
}

func (config *IConfig) SetValue() {
	jsonData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		log.Logger.Error(fmt.Sprintf("json.Marshal config error: %v", err))
		return
	}
	encryptData, err := sm4.Sm4CbcEncrypt(jsonData, constant.IconstantInstance.Sm4Key, constant.IconstantInstance.Sm4Iv)
	if err != nil {
		log.Logger.Error(fmt.Sprintf("sm4.Sm4CbcEncrypt error: %v", err))
		return
	}
	encodedString := hex.EncodeToString(encryptData)
	err = utils.WriteStringToFile(constant.IconstantInstance.SshConfigPath, encodedString)
	if err != nil {
		log.Logger.Error(fmt.Sprintf("utils.WriteStringToFile error: %v", err))
		return
	}

	log.Logger.Info("Config saved successfully.")
}

var (
	SshConfig IConfig
)

func init() {
	configStr, err := utils.ReadFileToString(constant.IconstantInstance.SshConfigPath)
	if err != nil {
		// 文件读取失败，可能是文件不存在，直接返回，后续流程会处理默认配置
		log.Logger.Warn(fmt.Sprintf("Failed to read config file, using default: %v", err))
		return
	}

	if configStr == "" {
		defaultData := `{"config": [],"is_dark": true,"is_english": true}`
		data := []byte(defaultData)

		encrypt, err := sm4.Sm4CbcEncrypt(data, constant.IconstantInstance.Sm4Key, constant.IconstantInstance.Sm4Iv)
		if err != nil {
			log.Logger.Error(fmt.Sprintf("Failed to encrypt default data: %v", err))
			return
		}

		err = utils.WriteStringToFile(constant.IconstantInstance.SshConfigPath, hex.EncodeToString(encrypt))
		if err != nil {
			log.Logger.Error(fmt.Sprintf("Failed to write default encrypted data: %v", err))
			return
		}

		err = json.Unmarshal(data, &SshConfig)
		if err != nil {
			return
		}
		log.Logger.Info("Initialized default config and stored it.")

	} else {

		data, err := hex.DecodeString(configStr)
		if err != nil {
			log.Logger.Error(fmt.Sprintf("failed to hex.DecodeString config: %v", err))
			return
		}

		decryptData, err := sm4.Sm4CbcDecrypt(data, constant.IconstantInstance.Sm4Key, constant.IconstantInstance.Sm4Iv)
		if err != nil {
			log.Logger.Error(fmt.Sprintf("failed to sm4.Sm4CbcDecrypt config: %v", err))
			err = os.Remove(constant.IconstantInstance.SshConfigPath)
			if err != nil {
				log.Logger.Error(fmt.Sprintf("failed to remove config file: %v", err))
				return
			}
			return
		}
		err = json.Unmarshal(decryptData, &SshConfig)
		if err != nil {
			log.Logger.Error(fmt.Sprintf("failed to json.Unmarshal config: %v", err))
			return
		}
		log.Logger.Info("Loaded and decrypted existing config.")
	}
}
