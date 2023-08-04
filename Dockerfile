FROM scratch
WORKDIR /project
COPY ./bin/protohack ./server
EXPOSE 8080 5000/udp
CMD [ "./server" ]
