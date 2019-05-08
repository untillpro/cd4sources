FROM golang:1.11.9

RUN mkdir /directcd
COPY directcd /directcd/directcd
COPY entrypoint.sh /directcd/entrypoint.sh
RUN chmod 777 /directcd/directcd
RUN chmod 777 /directcd/entrypoint.sh

ENTRYPOINT ["/directcd/entrypoint.sh"]
