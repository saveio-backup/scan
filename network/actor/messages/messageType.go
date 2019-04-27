/**
 * Description:
 * Author: LiYong Zhang
 * Create: 2019-04-01
 */
package messages

type UserDefineMsg interface {
	Desc(desc string)
}

type localMsgDemo struct{}

func (*localMsgDemo) Desc(desc string) {}
