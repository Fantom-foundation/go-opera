FROM pos-lachesis:latest as scratch


FROM alpine:latest

COPY --from=scratch /lachesis /

EXPOSE 55555 55556 55557

ENTRYPOINT ["/lachesis"]
