package main

import (
	"fmt"

	"github.com/silenceper/gogen/pkg/orm"
	"github.com/spf13/cobra"
)

var (
	exec    = "gogen"
	version = "v0.1"

	name    string
	ormFile string
)

var rootCmd = &cobra.Command{
	Use:   "gogen",
	Short: "gogen 是一个代码自动生成工具",
	Long:  `代码自动生成工具，包括golang app项目生成，orm生成`,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "show version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version)
	},
}

var genOrmCmd = &cobra.Command{
	Use:   "orm",
	Short: "根据 db.yml 文件自动生成orm",
	Run: func(cmd *cobra.Command, args []string) {
		orm.GenOrm(ormFile)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)

	rootCmd.AddCommand(genOrmCmd)
	genOrmCmd.PersistentFlags().StringVar(&ormFile, "orm", "db.yml", "数据库配置文件，用于生成orm")

}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		panic(err)
	}
}
