# build react
FROM node:20-alpine AS build_react_stage

RUN mkdir -p /home/react
WORKDIR /home/react

COPY client/package.json ./
RUN npm install

COPY client/ ./
RUN npm run build

# build go
FROM golang:1.21.6

WORKDIR /home/seek-tune

COPY server/go.mod server/go.sum ./
RUN go mod download

COPY server/ ./
ENV ENV=production

RUN mkdir -p static
COPY --from=build_react_stage /home/react/build static

RUN go build -o seek-tune.exe

EXPOSE 5000

CMD [ "/home/seek-tune/seek-tune.exe", "serve" ]
