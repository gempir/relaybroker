FROM scratch
ADD /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ADD relaybroker /
CMD ["/relaybroker"]
EXPOSE 3333