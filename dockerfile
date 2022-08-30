FROM filvenus/venus-buildenv AS buildenv

COPY ./go.mod ./venus-auth/go.mod
RUN export GOPROXY=https://goproxy.cn,direct && cd venus-auth   && go mod download 
COPY . ./venus-auth
RUN export GOPROXY=https://goproxy.cn,direct && cd venus-auth  && make static


FROM filvenus/venus-runtime

# copy the app from build env
COPY --from=buildenv  /go/venus-auth/venus-auth /app/venus-auth
COPY ./docker/script  /script

EXPOSE 8989

ENTRYPOINT ["/app/venus-auth","run"]
