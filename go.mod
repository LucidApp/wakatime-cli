module github.com/wakatime/wakatime-cli

go 1.18

require (
	github.com/Azure/go-ntlmssp v0.0.0-20211209120228-48547f28849e
	github.com/alecthomas/chroma v0.10.0
	github.com/danwakefield/fnmatch v0.0.0-20160403171240-cbb64ac3d964
	github.com/dlclark/regexp2 v1.4.0
	github.com/gandarez/go-olson-timezone v0.1.0
	github.com/juju/mutex v0.0.0-20180619145857-d21b13acf4bf
	github.com/kevinburke/ssh_config v1.2.0
	github.com/matishsiao/goInfo v0.0.0-20210923090445-da2e3fa8d45f
	github.com/mitchellh/go-homedir v1.1.0
	github.com/pkg/sftp v1.13.4
	github.com/sirupsen/logrus v1.8.1
	github.com/slongfield/pyfmt v0.0.0-20180124071345-020a7cb18bca
	github.com/spf13/cobra v1.3.0
	github.com/spf13/jwalterweatherman v1.1.0
	github.com/spf13/viper v1.10.1
	github.com/stretchr/testify v1.7.0
	github.com/yookoala/realpath v1.0.0
	go.etcd.io/bbolt v1.3.6
	golang.org/x/crypto v0.0.0-20220411220226-7b82a4e95df4
	gopkg.in/ini.v1 v1.66.3
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/fsnotify/fsnotify v1.5.1 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/juju/errors v0.0.0-20210818161939-5560c4c073ff // indirect
	github.com/juju/retry v0.0.0-20210818141810-5526f6f6ff07 // indirect
	github.com/juju/testing v0.0.0-20211215003918-77eb13d6cad2 // indirect
	github.com/juju/version v0.0.0-20210303051006-2015802527a8 // indirect
	github.com/kr/fs v0.1.0 // indirect
	github.com/magiconair/properties v1.8.5 // indirect
	github.com/mitchellh/mapstructure v1.4.3 // indirect
	github.com/pelletier/go-toml v1.9.4 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/spf13/afero v1.8.0 // indirect
	github.com/spf13/cast v1.4.1 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/subosito/gotenv v1.2.0 // indirect
	golang.org/x/sys v0.0.0-20220412211240-33da011f77ad // indirect
	golang.org/x/text v0.3.7 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
)

replace github.com/alecthomas/chroma => github.com/wakatime/chroma v0.10.0-wakatime.1

replace github.com/matishsiao/goInfo => github.com/wakatime/goInfo v0.1.0-wakatime.8
