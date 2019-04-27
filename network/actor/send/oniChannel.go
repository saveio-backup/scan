/**
 * Description:
 * Author: LiYong Zhang
 * Create: 2019-04-03
 */
package send

import (
	"github.com/ontio/ontology-eventbus/actor"
)

var ChannelMsgBusPid *actor.PID

func SetChannelMsgBusPid(conPid *actor.PID) {
	ChannelMsgBusPid = conPid
}
