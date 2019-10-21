package functions

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/dyrkin/bin"
	"github.com/dyrkin/zcl-go"
	"github.com/dyrkin/zcl-go/cluster"
	"github.com/dyrkin/zcl-go/frame"
	"github.com/dyrkin/zigbee-steward/coordinator"
	"github.com/dyrkin/znp-go"
	"reflect"
)

type commandClusterCache map[reflect.Type]uint8

type LocalSmartClusterFunctions struct {
	coordinator    *coordinator.Coordinator
	zcl            *zcl.Zcl
	commandIdCache map[cluster.ClusterId]commandClusterCache
}

func NewLocalSmartClusterFunctions(coordinator *coordinator.Coordinator, zcl *zcl.Zcl) *LocalSmartClusterFunctions {
	ccc := make(map[cluster.ClusterId]commandClusterCache)

	for ci, cl := range zcl.ClusterLibrary().Clusters() {
		ccc[ci] = make(map[reflect.Type]uint8)

		for cmdId, cmd := range cl.CommandDescriptors.Received {
			ccc[ci][reflect.TypeOf(cmd.Command)] = cmdId
		}
	}

	return &LocalSmartClusterFunctions{
		coordinator:    coordinator,
		zcl:            zcl,
		commandIdCache: ccc,
	}
}

func (f *LocalSmartClusterFunctions) IssueCommand(nwkAddress string, endpoint uint8, clusterId cluster.ClusterId, command interface{}) error {
	commandId, ok := f.lookupCommandId(clusterId, command)

	if !ok {
		return fmt.Errorf("failed to look up commandId on cluster [%d]", clusterId)
	}

	options := &znp.AfDataRequestOptions{}
	frm, err := frame.New().
		DisableDefaultResponse(false).
		FrameType(frame.FrameTypeLocal).
		Direction(frame.DirectionClientServer).
		CommandId(commandId).
		Command(command).
		Build()

	if err != nil {
		return err
	}

	response, err := f.coordinator.DataRequest(nwkAddress, endpoint, 1, uint16(clusterId), options, 15, bin.Encode(frm))
	if err == nil {
		zclIncomingMessage, err := f.zcl.ToZclIncomingMessage(response)
		if err == nil {
			zclCommand := zclIncomingMessage.Data.Command.(*cluster.DefaultResponseCommand)
			if zclCommand.Status != cluster.ZclStatusSuccess {
				return fmt.Errorf("unable to run command [%d] on cluster [%d]. Status: [%d]", commandId, clusterId, zclCommand.Status)
			}
			return nil
		} else {
			log.Errorf("Unsupported data response message:\n%s\n", func() string { return spew.Sdump(response) })
		}

	}
	return err
}

func (f *LocalSmartClusterFunctions) lookupCommandId(clusterId cluster.ClusterId, command interface{}) (uint8, bool) {
	commandCache, clusterPresent := f.commandIdCache[clusterId]

	if !clusterPresent {
		return 0, false
	}

	commandId, commandPresent := commandCache[reflect.TypeOf(command)]

	return commandId, commandPresent
}
