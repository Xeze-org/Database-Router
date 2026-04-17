package org.xeze.dbr;

import dbrouter.*;
import io.grpc.Channel;
import io.grpc.Grpc;
import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import io.grpc.TlsChannelCredentials;
import io.grpc.InsecureChannelCredentials;

import java.io.ByteArrayInputStream;
import java.io.File;
import java.io.InputStream;
import java.nio.file.Files;

public class XezeDbrClient implements AutoCloseable {

    private final ManagedChannel channel;

    public final HealthServiceGrpc.HealthServiceBlockingStub health;
    public final PostgresServiceGrpc.PostgresServiceBlockingStub postgres;
    public final MongoServiceGrpc.MongoServiceBlockingStub mongo;
    public final RedisServiceGrpc.RedisServiceBlockingStub redis;

    private XezeDbrClient(ManagedChannel channel) {
        this.channel = channel;
        this.health = HealthServiceGrpc.newBlockingStub(channel);
        this.postgres = PostgresServiceGrpc.newBlockingStub(channel);
        this.mongo = MongoServiceGrpc.newBlockingStub(channel);
        this.redis = RedisServiceGrpc.newBlockingStub(channel);
    }

    public static XezeDbrClient connect(Options opts) throws Exception {
        if (opts.insecure) {
            ManagedChannel channel = Grpc.newChannelBuilder(opts.host, InsecureChannelCredentials.create())
                    .build();
            return new XezeDbrClient(channel);
        }

        TlsChannelCredentials.Builder tlsBuilder = TlsChannelCredentials.newBuilder();

        if (opts.certData != null && opts.keyData != null) {
            try (InputStream certIs = new ByteArrayInputStream(opts.certData);
                 InputStream keyIs = new ByteArrayInputStream(opts.keyData)) {
                tlsBuilder.keyManager(certIs, keyIs);
            }
        } else if (opts.certFile != null && opts.keyFile != null) {
            tlsBuilder.keyManager(opts.certFile, opts.keyFile);
        } else {
            throw new IllegalArgumentException("Either certFile/keyFile or certData/keyData must be provided for secure connection.");
        }

        if (opts.caData != null) {
            try (InputStream caIs = new ByteArrayInputStream(opts.caData)) {
                tlsBuilder.trustManager(caIs);
            }
        } else if (opts.caFile != null) {
            tlsBuilder.trustManager(opts.caFile);
        }

        ManagedChannel channel = Grpc.newChannelBuilder(opts.host, tlsBuilder.build()).build();
        return new XezeDbrClient(channel);
    }

    @Override
    public void close() {
        if (channel != null && !channel.isShutdown()) {
            channel.shutdown();
        }
    }
}
