package rpc

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"

	"github.com/hacash/node/websocket"
)

func (api *DeprecatedApiService) webSocketHandler(ws *websocket.Conn) {
	var err error
	for {
		var reply string

		if err = websocket.Message.Receive(ws, &reply); err != nil {
			fmt.Println(err)
			ws.Close()
			break
		}

		if strings.HasPrefix(reply, "getblocks") {
			para := strings.Split(reply, " ")
			if len(para) != 2 {
				continue
			}

			start_height, e2 := strconv.ParseUint(para[1], 10, 0)
			if e2 != nil {
				continue
			}

			// read block data
			height_i := start_height
			resmaxsize := 1024 * 512
			totalsize := 0
			totaldatas := bytes.NewBuffer([]byte{0, 0, 0, 0})
			blockstore := api.blockchain.GetChainEngineKernel().StateRead().BlockStoreRead()
			for {
				_, blkbodybts, e2 := blockstore.ReadBlockBytesByHeight(height_i)
				if e2 != nil {
					fmt.Println(height_i, e2)
					break
				}

				if blkbodybts == nil {
					fmt.Println(height_i, "blkbodybts == nil")
					break // end
				}

				//fmt.Println(height_i, blkbodybts)
				totalsize += len(blkbodybts)
				totaldatas.Write(blkbodybts)
				height_i++
				if totalsize > resmaxsize {
					break
				}
			}

			fmt.Println("send block data by websocket, size:", totalsize, "blocknum", start_height, "~", height_i)
			if start_height == height_i {
				// end
				fmt.Println("send endblocks.")
				websocket.Message.Send(ws, []byte("endblocks"))
				ws.Close()
				break
			}

			// send results
			totalbytes := totaldatas.Bytes()
			binary.BigEndian.PutUint32(totalbytes[0:4], uint32(len(totalbytes))-4) // put data len
			if err = websocket.Message.Send(ws, totalbytes); err != nil {
				fmt.Println(err)
				ws.Close()
				break
			}
		}
	}
}
