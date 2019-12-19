FROM scratch

COPY ./server-static /server

# as no other directory exists
ENV TMPDIR=/

ENTRYPOINT ["/server"]
