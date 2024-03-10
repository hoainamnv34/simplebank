package main

import (
	"context"
	"database/sql"
	"net"
	"net/http"
	"os"
	"simplebank/api"
	db "simplebank/db/sqlc"
	_ "simplebank/doc/statik"
	"simplebank/gapi"
	"simplebank/pb"
	"simplebank/util"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	_ "github.com/lib/pq"
	"github.com/rakyll/statik/fs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
)

func main() {

	config, err := util.LoadConfig(".")

	if err != nil {
		log.Fatal().Msg("cannot load config")
	}

	if config.Environment == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	conn, err := sql.Open(config.DBDriver, config.DBSource)

	if err != nil {
		log.Fatal().Msg("connot connect to db")
	}

	runDBMigration(config.MigrationURL, config.DBSource)

	store := db.NewStore(conn)

	go runGatewayServer(config, store)
	runGrpcServer(config, store)

}

func runGrpcServer(config util.Config, store db.Store) {
	server, err := gapi.NewServer(config, store)

	if err != nil {
		log.Fatal().Msg("connot create server")
	}

	gprcLogger := grpc.UnaryInterceptor(gapi.GrpcLogger)
	grpcServer := grpc.NewServer(gprcLogger)
	pb.RegisterSimpleBankServer(grpcServer, server)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", config.GRPCServerAddress)

	if err != nil {
		log.Fatal().Msg("connot create listener")
	}

	log.Info().Msgf("Start gRPC server at %s", listener.Addr().String())
	err = grpcServer.Serve(listener)

	if err != nil {
		log.Fatal().Msg("connot start gRPC server")
	}
}

func runGinServer(config util.Config, store db.Store) {
	server, err := api.NewServer(config, store)

	if err != nil {
		log.Fatal().Msg("connot create server")
	}
	err = server.Start(config.HTTPServerAddress)

	if err != nil {
		log.Fatal().Msg("connot start server")
	}
}

func runGatewayServer(
	// ctx context.Context,
	// waitGroup *errgroup.Group,
	config util.Config,
	store db.Store,
	// taskDistributor worker.TaskDistributor,
) {
	server, err := gapi.NewServer(config, store)
	if err != nil {
		log.Fatal().Msg("cannot create server")
	}

	jsonOption := runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			UseProtoNames: true,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	})

	grpcMux := runtime.NewServeMux(jsonOption)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = pb.RegisterSimpleBankHandlerServer(ctx, grpcMux, server)
	if err != nil {
		log.Fatal().Msg("cannot register handler server")
	}

	mux := http.NewServeMux()
	mux.Handle("/", grpcMux)

	statikFS, err := fs.New()
	if err != nil {
		log.Fatal().Msg("cannot create statik fs")
	}

	swaggerHandler := http.StripPrefix("/swagger/", http.FileServer(statikFS))
	mux.Handle("/swagger/", swaggerHandler)

	listener, err := net.Listen("tcp", config.HTTPServerAddress)

	if err != nil {
		log.Fatal().Msg("connot create listener")
	}

	log.Info().Msgf("Start HTTP gateway server at %s", listener.Addr().String())

	handler := gapi.HttpLogger(mux)
	err = http.Serve(listener, handler)

	if err != nil {
		log.Fatal().Msg("connot start HTTP gateway server")
	}

	// httpServer := &http.Server{
	// 	Handler: gapi.HttpLogger(mux),
	// 	Addr:    config.HTTPServerAddress,
	// }

	// waitGroup.Go(func() error {
	// 	log.Info().Msgf("start HTTP gateway server at %s", httpServer.Addr)
	// 	err = httpServer.ListenAndServe()
	// 	if err != nil {
	// 		if errors.Is(err, http.ErrServerClosed) {
	// 			return nil
	// 		}
	// 		log.Error().Err(err).Msg("HTTP gateway server failed to serve")
	// 		return err
	// 	}
	// 	return nil
	// })

	// waitGroup.Go(func() error {
	// 	<-ctx.Done()
	// 	log.Info().Msg("graceful shutdown HTTP gateway server")

	// 	err := httpServer.Shutdown(context.Background())
	// 	if err != nil {
	// 		log.Error().Err(err).Msg("failed to shutdown HTTP gateway server")
	// 		return err
	// 	}

	// 	log.Info().Msg("HTTP gateway server is stopped")
	// 	return nil
	// })
}

func runDBMigration(migrationURL string, dbSource string) {
	migration, err := migrate.New(migrationURL, dbSource)
	if err != nil {
		log.Fatal().Msg("cannot create new migrate instance")
	}

	if err = migration.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal().Msg("failed to run migrate up")
	}

	log.Info().Msg("db migrated successfully")
}
