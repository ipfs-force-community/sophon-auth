FROM filvenus/venus-buildenv AS buildenv

RUN git clone https://github.com/filecoin-project/venus-auth.git --depth 1 
RUN export GOPROXY=https://goproxy.cn && cd venus-auth  && make linux


FROM filvenus/venus-runtime

# DIR for app
WORKDIR /app

# copy the app from build env
COPY --from=buildenv  /go/venus-auth/venus-auth /app/venus-auth
COPY ./docker/script  /script

EXPOSE 8989

ENTRYPOINT ["/app/venus-auth","run"]


