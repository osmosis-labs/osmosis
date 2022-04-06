FROM informalsystems/hermes:0.12.0 AS hermes-builder

FROM debian:buster-slim
USER root

COPY --chown=0:0 --from=hermes-builder /usr/lib/x86_64-linux-gnu/libssl.so.1.1 /usr/lib/x86_64-linux-gnu/libssl.so.1.1
COPY --chown=0:0 --from=hermes-builder /usr/lib/x86_64-linux-gnu/libcrypto.so.1.1 /usr/lib/x86_64-linux-gnu/libcrypto.so.1.1
COPY --from=hermes-builder /usr/bin/hermes /usr/local/bin/
RUN chmod +x /usr/local/bin/hermes

EXPOSE 3031
ENTRYPOINT ["hermes", "start"]