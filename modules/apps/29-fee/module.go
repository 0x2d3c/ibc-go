package fee

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	"cosmossdk.io/client/v2/autocli"
	"cosmossdk.io/core/appmodule"
	coreregistry "cosmossdk.io/core/registry"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/cosmos/ibc-go/v9/modules/apps/29-fee/client/cli"
	"github.com/cosmos/ibc-go/v9/modules/apps/29-fee/keeper"
	"github.com/cosmos/ibc-go/v9/modules/apps/29-fee/types"
)

var (
	_ appmodule.AppModule             = (*AppModule)(nil)
	_ appmodule.HasConsensusVersion   = (*AppModule)(nil)
	_ appmodule.HasAminoCodec         = (*AppModule)(nil)
	_ appmodule.HasRegisterInterfaces = (*AppModule)(nil)
	_ appmodule.HasMigrations         = (*AppModule)(nil)

	_ module.AppModule  = (*AppModule)(nil)
	_ module.HasGenesis = (*AppModule)(nil)

	_ autocli.HasCustomTxCommand    = (*AppModule)(nil)
	_ autocli.HasCustomQueryCommand = (*AppModule)(nil)
)

// AppModule represents the AppModule for this module
type AppModule struct {
	cdc    codec.Codec
	keeper keeper.Keeper
}

// NewAppModule creates a new 29-fee module
func NewAppModule(cdc codec.Codec, k keeper.Keeper) AppModule {
	return AppModule{
		cdc:    cdc,
		keeper: k,
	}
}

// Name implements AppModuleBasic interface
func (AppModule) Name() string {
	return types.ModuleName
}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (AppModule) IsOnePerModuleType() {}

// IsAppModule implements the appmodule.AppModule interface.
func (AppModule) IsAppModule() {}

// RegisterLegacyAminoCodec implements AppModule interface
func (AppModule) RegisterLegacyAminoCodec(cdc coreregistry.AminoRegistrar) {
	types.RegisterLegacyAminoCodec(cdc)
}

// RegisterInterfaces registers module concrete types into protobuf Any.
func (AppModule) RegisterInterfaces(registry coreregistry.InterfaceRegistrar) {
	types.RegisterInterfaces(registry)
}

// DefaultGenesis returns default genesis state as raw bytes for the ibc
// 29-fee module.
func (am AppModule) DefaultGenesis() json.RawMessage {
	return am.cdc.MustMarshalJSON(types.DefaultGenesisState())
}

// ValidateGenesis performs genesis state validation for the 29-fee module.
func (am AppModule) ValidateGenesis(bz json.RawMessage) error {
	var gs types.GenesisState
	if err := am.cdc.UnmarshalJSON(bz, &gs); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}

	return gs.Validate()
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for ics29 fee module.
func (AppModule) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	err := types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx))
	if err != nil {
		panic(err)
	}
}

// GetTxCmd implements AppModule interface
func (AppModule) GetTxCmd() *cobra.Command {
	return cli.NewTxCmd()
}

// GetQueryCmd implements AppModule interface
func (AppModule) GetQueryCmd() *cobra.Command {
	return cli.GetQueryCmd()
}

func (am AppModule) RegisterMigrations(registrar appmodule.MigrationRegistrar) error {
	m := keeper.NewMigrator(am.keeper)
	if err := registrar.Register(types.ModuleName, 1, m.Migrate1to2); err != nil {
		return fmt.Errorf("failed to migrate ibc-fee module from version 1 to 2 (refund leftover fees): %v", err)
	}
	return nil
}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg grpc.ServiceRegistrar) error {
	types.RegisterMsgServer(cfg, am.keeper)
	types.RegisterQueryServer(cfg, am.keeper)
	return nil
}

// InitGenesis performs genesis initialization for the ibc-29-fee module. It returns
// no validator updates.
func (am AppModule) InitGenesis(ctx context.Context, data json.RawMessage) error {
	var genesisState types.GenesisState
	am.cdc.MustUnmarshalJSON(data, &genesisState)
	am.keeper.InitGenesis(ctx, genesisState)
	return nil
}

// ExportGenesis returns the exported genesis state as raw bytes for the ibc-29-fee
// module.
func (am AppModule) ExportGenesis(ctx context.Context) (json.RawMessage, error) {
	gs := am.keeper.ExportGenesis(ctx)
	return am.cdc.MarshalJSON(gs)
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return 2 }
