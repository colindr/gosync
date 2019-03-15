package cmd

import (
	"encoding/gob"
	"errors"
	"fmt"
	"github.com/colindr/gotests/gosync/transfer"
	"github.com/google/uuid"
	"net"
	"strconv"
	"strings"
	"github.com/colindr/gotests/gosync/request"
	"os"

	"github.com/spf13/cobra"
)

var host string
var port int
var configFile string

func init() {
}

var rootCmd = &cobra.Command{
	Use:   "gosync",
	Short: "gosync syncs files from gosyncd",
	Long:  `An exercise in golang to sync files`,
	Args: func(cmd *cobra.Command, args []string) error {
		// pass
		if len(args)!= 2 {
			return errors.New("requires exactly 2 args")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		// Perform a sync
		source := args[0]
		dest := args[1]
		req, err := NewRequestFromSourceAndDestination(source, dest)

		if err != nil {
			fmt.Println(err)
			return
		}

		if err := InitiateSync(req); err != nil {
			fmt.Println(err)
			return
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}


func NewRequestFromSourceAndDestination(source string, dest string) (*request.Request, error) {
	source_parts := strings.Split(source, ":")
	dest_parts := strings.Split(dest, ":")

	if len(source_parts) > 3 {
		return nil, errors.New(fmt.Sprintf("unknown sync source format: %v", source))
	}

	if len(dest_parts) > 3 {
		return nil, errors.New(fmt.Sprintf("unknown sync source format: %v", source))
	}

	if (len(source_parts) > 1) && (len(dest_parts) > 1) {
		return nil, errors.New(fmt.Sprintf("only one of source or destination can specify a host"))
	}

	var transferType request.Type
	var path string
	var destination string
	var port int
	var host string

	if (len(source_parts) == 1) && (len(dest_parts) == 1) {
		transferType = request.Local
		path = source
		destination = dest
		host = ""
		port = 0
	} else if len(source_parts) == 1 {
		transferType = request.Outgoing
		path = source
		host = dest_parts[0]
		if len (dest_parts) == 2{
			if p, err := strconv.Atoi(dest_parts[1]); err != nil {
				return nil, errors.New(fmt.Sprintf("unparsable port number: %v", dest_parts[1] ))
			} else{
				port = p
			}
			destination = dest_parts[2]
		} else {
			port = 4200 // TODO: default tcp port
			destination = dest_parts[1]
		}
	} else {
		transferType = request.Incoming
		path = dest
		host = source_parts[0]
		if len (source_parts) == 2{
			if p, err := strconv.Atoi(source_parts[1]); err != nil {
				return nil, errors.New(fmt.Sprintf("unparsable port number: %v", dest_parts[1] ))
			} else{
				port = p
			}
			destination = source_parts[2]
		} else {
			port = 4200 // TODO: default tcp port
			destination = source_parts[1]
		}
	}

	return &request.Request{
		RequestID: uuid.New(),

		Host: host,
		Port: port,

		Type: transferType,

		Path: path,
		Destination: destination,

		FollowLinks: false,

	}, nil

}

func InitiateSync(req *request.Request) error {
	if req.Type == request.Local {
		return transfer.SyncLocal(req)
	}

	addr := fmt.Sprintf("%s:%v", req.Host, req.Port)
	fmt.Println("Connecting to", addr)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}

	encoder := gob.NewEncoder(conn)

	if err := encoder.Encode(req); err != nil {
		return errors.New(fmt.Sprintln("Error encoding transfer request:", err))
	}

	decoder := gob.NewDecoder(conn)

	resp := &request.RequestResponse{}
	if err := decoder.Decode(resp); err != nil {
		return errors.New(fmt.Sprintln("Error encoding transfer request:", err))
	}

	if !resp.Accepted {
		return errors.New(fmt.Sprintln("Transfer request rejected:", resp.Reason))
	}

	return transfer.Sync(&conn, req, resp)

}