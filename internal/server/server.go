package server

import (
	"context"
	"errors"
	"io"
	"sync"

	"github.com/sirupsen/logrus"
	"go.uber.org/atomic"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"

	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/engine"
	gameserverpb "github.com/c4t-but-s4d/ctfcup-2024-igra/proto/go/gameserver"
)

func New(game *Game, factory engine.Factory, round int64) *GameServer {
	return &GameServer{
		factory:    factory,
		game:       game,
		numStreams: atomic.NewInt64(0),
		round:      round,
	}
}

type GameServer struct {
	gameserverpb.UnimplementedGameServerServiceServer

	factory      engine.Factory
	numStreams   *atomic.Int64
	game         *Game
	round        int64
	mu           sync.Mutex
	lastResponse *gameserverpb.InventoryResponse
}

func (g *GameServer) Ping(context.Context, *gameserverpb.PingRequest) (*gameserverpb.PingResponse, error) {
	return &gameserverpb.PingResponse{}, nil
}

func (g *GameServer) ProcessEvent(stream gameserverpb.GameServerService_ProcessEventServer) error {
	defer g.numStreams.Dec()
	if g.numStreams.Inc() > 1 {
		return status.Error(codes.ResourceExhausted, "only one client connection allowed")
	}

	p, _ := peer.FromContext(stream.Context())
	if p == nil {
		return status.Error(codes.FailedPrecondition, "failed to get peer info")
	}
	logrus.Infof("new connection from %v", p.Addr)

	eng, err := g.factory()
	if err != nil {
		return status.Errorf(codes.Internal, "creating engine: %v", err)
	}

	g.game.setEngine(eng)
	defer g.game.resetEngine()

	if err := stream.Send(&gameserverpb.ServerEvent{
		Event: &gameserverpb.ServerEvent_Snapshot{
			Snapshot: eng.StartSnapshot.ToProto(),
		},
	}); err != nil {
		return status.Errorf(codes.Internal, "failed to send start snapshot: %v", err)
	}

	for {
		req, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			logrus.Info("client disconnected")
			return nil
		}
		if err != nil {
			return status.Errorf(codes.Internal, "failed to read from stream: %v", err)
		}
		logrus.Debugf("received event: %v", req)

		if req.Event == nil {
			return status.Error(codes.InvalidArgument, "event is nil")
		}
		if err := eng.ValidateChecksum(req.Checksum); err != nil {
			g.game.WasCheating = true
			return status.Errorf(codes.InvalidArgument, "invalid checksum: %v", err)
		}

		if npc := eng.ActiveNPC(); npc != nil {
			event := &gameserverpb.ServerEvent{Event: &gameserverpb.ServerEvent_GameEvent{
				GameEvent: &gameserverpb.GameEvent{
					DialogState: npc.Dialog.State().ToProto(),
				},
			}}
			logrus.Debugf("sending event: %v", req)
			if err := stream.Send(event); err != nil {
				return status.Errorf(codes.Internal, "failed to send game event: %v", err)
			}
		}
		if arc := eng.ActiveArcade(); arc != nil {
			event := &gameserverpb.ServerEvent{Event: &gameserverpb.ServerEvent_GameEvent{
				GameEvent: &gameserverpb.GameEvent{
					ArcadeState: arc.Game.State().ToProto(),
				},
			}}
			logrus.Debugf("sending event: %v", req)
			if err := stream.Send(event); err != nil {
				return status.Errorf(codes.Internal, "failed to send game event: %v", err)
			}
		}

		if err := g.game.processEvent(req.Event); err != nil {
			return status.Errorf(codes.Internal, "processing event: %v", err)
		}

		g.updateLastResponse()
		g.updateIsWin()
	}
}

func (g *GameServer) updateLastResponse() {
	g.mu.Lock()
	defer g.mu.Unlock()

	eng := g.game.getEngine()
	if eng != nil {
		g.lastResponse = &gameserverpb.InventoryResponse{Inventory: eng.Player.Inventory.ToProto(), Round: g.round}
	} else if g.lastResponse == nil {
		g.lastResponse = &gameserverpb.InventoryResponse{Round: g.round}
	}
}

func (g *GameServer) updateIsWin() {
	g.mu.Lock()
	defer g.mu.Unlock()

	eng := g.game.getEngine()
	if eng != nil {
		g.game.IsWin = eng.IsWin
	}
}

func (g *GameServer) GetInventory(context.Context, *gameserverpb.InventoryRequest) (*gameserverpb.InventoryResponse, error) {
	g.updateLastResponse()

	return g.lastResponse, nil
}
