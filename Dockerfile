FROM golang:1.11.9

RUN mkdir /directcd
COPY directcd /directcd/directcd
RUN chmod 777 /directcd/directcd

ENTRYPOINT ["/directcd/directcd"]
