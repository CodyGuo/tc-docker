package tc

import (
	"errors"
	"fmt"
	"strings"

	"github.com/CodyGuo/glog"
	"github.com/CodyGuo/tc-docker/pkg/command"
)

var (
	ErrTcNotFound = errors.New("RTNETLINK answers: No such file or directory")
)

func SetTcRate(veth, rate, ceil string) error {
	delRootHandleCmd := fmt.Sprintf("/usr/sbin/tc qdisc del dev %s root", veth)
	glog.Debug(delRootHandleCmd)
	out, err := command.CombinedOutput(delRootHandleCmd)
	if err != nil {
		if strings.TrimSpace(string(out)) != ErrTcNotFound.Error() {
			return fmt.Errorf("out: %s, error: %w", out, err)
		}
	}
	addRootHandleCmd := fmt.Sprintf("/usr/sbin/tc qdisc add dev %s root handle 1a1a: htb default 1", veth)
	glog.Debug(addRootHandleCmd)
	out, err = command.CombinedOutput(addRootHandleCmd)
	if err != nil {
		return fmt.Errorf("out: %s, error: %w", out, err)
	}
	addClassRateCmd := fmt.Sprintf("/usr/sbin/tc class add dev %s parent 1a1a: classid 1a1a:1 htb rate %s ceil %s prio 2", veth, rate, ceil)
	glog.Debug(addClassRateCmd)
	out, err = command.CombinedOutput(addClassRateCmd)
	if err != nil {
		return fmt.Errorf("out: %s, error: %w", out, err)
	}
	addSfqHandleCmd := fmt.Sprintf("/usr/sbin/tc qdisc add dev %s parent 1a1a:1 handle 10: sfq perturb 10", veth)
	glog.Debug(addSfqHandleCmd)
	out, err = command.CombinedOutput(addSfqHandleCmd)
	if err != nil {
		return fmt.Errorf("out: %s, error: %w", out, err)
	}
	addFilterCmd := fmt.Sprintf("/usr/sbin/tc filter add dev %s protocol ip parent 1a1a: prio 2 u32 match ip src 0.0.0.0/0 match ip dst 0.0.0.0/0 flowid 1a1a:1", veth)
	glog.Debug(addFilterCmd)
	out, err = command.CombinedOutput(addFilterCmd)
	if err != nil {
		return fmt.Errorf("out: %s, error: %w", out, err)
	}
	return nil
}
