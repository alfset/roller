package relayer

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
)

func (r *Relayer) LoadChannels() (string, string, error) {
	output, err := utils.ExecBashCommandWithStdout(r.queryChannelsRollappCmd())
	if err != nil {
		return "", "", err
	}

	if output.Len() == 0 {
		return "", "", nil
	}

	// While there are JSON objects in the stream...
	var outputStruct RollappQueryResult
	dec := json.NewDecoder(&output)
	for dec.More() {
		err = dec.Decode(&outputStruct)
		if err != nil {
			return "", "", fmt.Errorf("error while decoding JSON: %v", err)
		}
		continue
	}

	if outputStruct.State != "STATE_OPEN" {
		r.logger.Printf("channel %s is not STATE_OPEN (%s)", outputStruct.ChannelID, outputStruct.State)
		return "", "", nil
	}

	// Check if the channel is open on the hub
	var res HubQueryResult
	outputHub, err := utils.ExecBashCommandWithStdout(r.queryChannelsHubCmd(outputStruct.Counterparty.ChannelID))
	if err != nil {
		return "", "", err
	}
	err = json.Unmarshal(outputHub.Bytes(), &res)
	if err != nil {
		return "", "", err
	}

	if res.Channel.State != "STATE_OPEN" {
		r.logger.Printf("channel %s is STATE_OPEN on the rollapp, but channel %s is %s on the hub",
			outputStruct.ChannelID, outputStruct.Counterparty.ChannelID, res.Channel.State,
		)
		return "", "", nil
	}

	r.SrcChannel = outputStruct.ChannelID
	r.DstChannel = outputStruct.Counterparty.ChannelID
	return r.SrcChannel, r.DstChannel, nil
}

func (r *Relayer) queryChannelsRollappCmd() *exec.Cmd {
	args := []string{"q", "channels", r.RollappID}
	args = append(args, "--home", filepath.Join(r.Home, consts.ConfigDirName.Relayer))
	return exec.Command(consts.Executables.Relayer, args...)
}

func (r *Relayer) queryChannelsHubCmd(channelID string) *exec.Cmd {
	args := []string{"q", "channel", r.HubID, channelID, "transfer"}
	args = append(args, "--home", filepath.Join(r.Home, consts.ConfigDirName.Relayer))
	return exec.Command(consts.Executables.Relayer, args...)
}

func (r *Relayer) ChannelReady() bool {
	return r.SrcChannel != "" && r.DstChannel != ""
}

type Counterparty struct {
	PortID    string `json:"port_id"`
	ChannelID string `json:"channel_id"`
}

type Output struct {
	State          string       `json:"state"`
	Ordering       string       `json:"ordering"`
	Counterparty   Counterparty `json:"counterparty"`
	ConnectionHops []string     `json:"connection_hops"`
	Version        string       `json:"version"`
}

type ProofHeight struct {
	RevNumber string `json:"revision_number"`
	RevHeight string `json:"revision_height"`
}
type HubQueryResult struct {
	Channel     Output      `json:"channel"`
	Proof       string      `json:"proof"`
	ProofHeight ProofHeight `json:"proof_height"`
}

type RollappQueryResult struct {
	Output
	PortID    string `json:"port_id"`
	ChannelID string `json:"channel_id"`
}
