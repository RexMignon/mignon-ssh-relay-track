package constant

import "encoding/hex"

type (
	IConstant struct {
		Sm4Key        []byte
		Sm4Iv         []byte
		SshConfigPath string
		LoggerPath    string
	}
)

var (
	IconstantInstance IConstant
)

func init() {
	sm4Key := "68e4da8059897460a2e2eef0f8f04aea"
	sm4Iv := "aec61a049d064ea7d0dc0f5ed070d42e"
	key, _ := hex.DecodeString(sm4Key)
	iv, _ := hex.DecodeString(sm4Iv)
	IconstantInstance = IConstant{
		key,
		iv,
		"./resources/config/mignon_ssh_config.rex",
		"./resources/log/app.log",
	}
}
