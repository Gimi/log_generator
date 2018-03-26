FROM scratch
LABEL maintainer="Zhimin (Gimi) Liang <Gimi @ github>"

COPY genlog /
# nobody:nobody
USER 65534:65534
ENTRYPOINT ["/genlog"]
