// Author: xr@meith.com
// Date: 2018-03-28
// Desc: A simple wrapper for retrieving all the blocks

package commands

import (
	"fmt"
	"io"
	"strings"
	"errors"

	"gx/ipfs/QmTVDM4LCSUMFNQzbDLL9zQwp8usE6QHymFdh3h8vL9v6b/go-ipfs-blockstore"
	bserv "github.com/ipfs/go-ipfs/blockservice"
	"github.com/ipfs/go-ipfs/pin"

	"github.com/ipfs/go-ipfs/repo/fsrepo"
	cmds "github.com/ipfs/go-ipfs/commands"
	"gx/ipfs/QmceUdzxkimdYsgtX733uNgzf1DLHyBKN6ehGSp85ayppM/go-ipfs-cmdkit"
	"gx/ipfs/QmcZfnkapfECQGcLZaf9B79NRg7cRa9EnZh4LSbkCzwNvY/go-cid"

	"github.com/ipfs/go-ipfs/merkledag"
	ipld "gx/ipfs/Qme5bWv7wtjUNGsK2BNGVUFPKiuxWrsqrtvYwCLRw8YFES/go-ipld-format"

	"context"
)

type objectInfo struct {
	Cid       *cid.Cid
	Type      string
	TotalSize uint64
	Pinned    bool
}

type objectInfos []objectInfo

// Implements the interface for sorting
func (ois objectInfos) Len() int {
	return len(ois)
}

func (ois objectInfos) Swap(i, j int) {
	ois[i], ois[j] = ois[j], ois[i]
}

func (ois objectInfos) Less(i, j int) bool {
	if ois[i].Type == "unknown" {
		if ois[j].Type != "unknown" {
			return false
		}

		if ois[i].Pinned && !ois[j].Pinned {
			return true
		}

		return ois[i].TotalSize > ois[j].TotalSize
	}

	if ois[j].Type == "unknown" {
		return true
	}

	if ois[i].Pinned && !ois[j].Pinned {
		return true
	}

	return ois[i].TotalSize > ois[j].TotalSize
}

func RunCommand() (io.Reader, error) {
	p, err := fsrepo.BestKnownPath()
	if err != nil {
		return nil, errors.New("Failed to get best known path")
	}

	fmt.Println("running command")
	r, err := fsrepo.Open(p)
	if err != nil {
		return nil, errors.New("Failed to open best known path, have you turned your daemon off?")
	}

	bs := blockstore.NewBlockstore(r.Datastore())
	dag := merkledag.NewDAGService(bserv.New(bs, nil))

	pinner, err := pin.LoadPinner(r.Datastore(), dag, dag)
	if err != nil {
		return nil, errors.New("Failed to load pinner")
	}

	// print lost-pins
	maybeLost, err := findMaybeLostPins(bs, dag, pinner)
	if err != nil {
		return nil, errors.New("Failed to find lost pins")
	}

	out := ""
	for _, c := range maybeLost {
		out = fmt.Sprintln(out, c)
	}

	return strings.NewReader(out), nil
}

func findMaybeLostPins(blks blockstore.Blockstore, dag ipld.DAGService, pinner pin.Pinner) ([]*cid.Cid, error) {
	pins := cid.NewSet()

	for _, reck := range pinner.RecursiveKeys() {
		pins.Add(reck)
	}

	for _, dirk := range pinner.DirectKeys() {
		pins.Add(dirk)
	}

	seen := cid.NewSet()

	kchan, err := blks.AllKeysChan(context.Background())
	if err != nil {
		return nil, err
	}

	missing := cid.NewSet()
	for c := range kchan {
		err := processObject(dag, c, seen, pins, missing)
		if err != nil {
			return nil, err
		}
	}

	return missing.Keys(), nil
}

func processObject(dag ipld.DAGService, c *cid.Cid, seen, pinned, missing *cid.Set) error {
	if seen.Has(c) {
		return nil
	}

	seen.Add(c)

	_, err := dag.Get(context.Background(), c)
	if err != nil {
		return nil
	}

	// TODO
	return nil
}

type SeeAllOutput struct {
	Block []byte
	Count int
}

var SeeAllCmd = &cmds.Command{
	Helptext: cmdkit.HelpText{
		Tagline:          "Show all the blocks within ipfs",
		ShortDescription: "Returns all the blocks",
	},

	PreRun: func(req cmds.Request) error {
		_, err := RunCommand()
		return err
	},
	Run: func(req cmds.Request, res cmds.Response) {
		res.SetOutput(&SeeAllOutput{
			Block: []byte(string("rixon")),
			Count: 1,
		})
	},
	Marshalers: cmds.MarshalerMap{
		cmds.Text: func(res cmds.Response) (io.Reader, error) {
			v, err := unwrapOutput(res.Output())
			if err != nil {
				return nil, err
			}

			seeAll, _ := v.(*SeeAllOutput)
			commitTxt := ""
			out := fmt.Sprintln("result is: ", commitTxt, string(seeAll.Block), seeAll.Count)

			return strings.NewReader(out), nil
		},
	},
	Type: SeeAllOutput{},
}
