FROM golang:1.15.2-alpine3.12
RUN apk add git
RUN mkdir /app

ADD . /app
ADD ./db /go/src/apexrand/db
ADD ./random /go/src/apexrand/random
ADD ./qrw /go/src/apexrand/qrw
ADD ./file /go/src/apexrand/file
ADD ./images /go/src/apexrand/images

WORKDIR /app
#RUN go get /C/Users/Randy/Dropbox/Prog/GoProj/src/qrw
RUN go get github.com/go-sql-driver/mysql
RUN go build -o apex .
#HEALTHCHECK CMD curl --fail http://mysql_apex:3306/ || exit 1

#COPY wait-for-it.sh wait-for-it.sh 
#RUN chmod +x wait-for-it.sh
#ENTRYPOINT [ "/bin/bash", "-c" ]
#CMD ["./wait-for-it.sh" , "[ENTER YOUR ENDPOINT HERE]" , "--strict" , "--timeout=300" , "--" , "YOUR REAL START COMMAND"]
#CMD ["./wait-for-it.sh" , "mysql_apex:3306" , "--strict" , "--timeout=300" , "--" , "app/apex"]


CMD ["/app/apex"]