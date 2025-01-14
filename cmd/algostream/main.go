// Copyright (C) 2022 AlgoNode Org.
//
// algostreamer is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// algostreamer is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with algostreamer.  If not, see <https://www.gnu.org/licenses/>.

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	//load config
	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %s", err)
		return
	}

	//make us a nice cancellable context
	//set Ctrl-C as the cancell trigger
	ctx, cf := context.WithCancel(context.Background())
	defer cf()
	{
		cancelCh := make(chan os.Signal, 1)
		signal.Notify(cancelCh, syscall.SIGTERM, syscall.SIGINT)
		go func() {
			<-cancelCh
			fmt.Fprintf(os.Stderr, "Stopping streamer.\n")
			cf()
		}()
	}

	//spawn a block stream fetcher that never fails
	blocks, err := algodStream(ctx, &cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting algod stream: %s", err)
		return
	}

	//spawn a redis pusher
	err = redisPusher(ctx, &cfg, blocks)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error setting up redis: %s", err)
		return
	}

	//Wait for the end of the Algoverse
	<-ctx.Done()

}
