/**
 * Description:
 * Author: LiYong Zhang
 * Create: 2019-04-03
 */
package send

import (
	"github.com/ontio/ontology-eventbus/actor"
)

var TrackerMsgBusPid *actor.PID

func SetTrackerMsgBusPid(conPid *actor.PID) {
	TrackerMsgBusPid = conPid
}
