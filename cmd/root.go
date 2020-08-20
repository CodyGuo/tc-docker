package cmd

import (
	"fmt"

	"github.com/CodyGuo/glog"
	"github.com/CodyGuo/tc-docker/global"
	"github.com/CodyGuo/tc-docker/internal/docker"
	"github.com/CodyGuo/tc-docker/internal/tc"
	"github.com/spf13/cobra"
)

var debug bool

func init() {
	rootCmd.Flags().BoolVarP(&debug, "debug", "d", false, "set logger debug")
}

var rootCmd = &cobra.Command{
	Use:   "",
	Short: "",
	Long:  "",
	PreRun: func(cmd *cobra.Command, args []string) {
		if debug {
			glog.SetLevel(glog.DEBUG)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		c := docker.NewContainer(global.Ctx, global.DockerClient)
		containers, err := c.GetRunningList()
		if err != nil {
			glog.Fatal(err)
		}
		for _, container := range containers {
			err := tc.SetTcRate(container.Veth, container.TcRate, container.TcCeil)
			if err != nil {
				glog.Errorf("SetTcRate failed, container: %s, id: %s, error: %v", container.Name, container.ID, err)
				continue
			}
			glog.Infof("SetTcRate success, container: %s, id: %s, veth: %s, rate: %s, ceil: %s",
				container.Name, container.ID, container.Veth, container.TcRate, container.TcCeil)
		}

		startErr := c.EventStart(func(container docker.Container) error {
			err := tc.SetTcRate(container.Veth, container.TcRate, container.TcCeil)
			if err != nil {
				return fmt.Errorf("SetTcRate failed, container: %s, id: %s, error: %v", container.Name, container.ID, err)
			}
			glog.Infof("AutoDiscover SetTcRate success, container: %s, id: %s, veth: %s, rate: %s, ceil: %s",
				container.Name, container.ID, container.Veth, container.TcRate, container.TcCeil)
			return nil
		})
		dieErr := c.EventDie(func(container docker.Container) error {
			glog.Infof("container stopped, name: %s, id: %s", container.Name, container.ID)
			return nil
		})
		for {
			select {
			case err := <-startErr:
				glog.Errorf("EventStart error: %v", err)
			case err := <-dieErr:
				glog.Errorf("EventDie error: %v", err)
			}
		}
	},
}

func Execute() error {
	return rootCmd.Execute()
}
