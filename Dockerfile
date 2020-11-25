FROM archlinux

# Setup in your Spotify Developer Account
ENV SPOTIFY_CLIENT_ID=""
ENV SPOTIFY_CLIENT_SECRET=""
# Callback URI after auth, 
# ie myapp://spotify_auth_callback
ENV SPOTIFY_AUTH_REDIRECT_URI=""

EXPOSE $PORT
ARG SRV_DIR=/srv/spotify_auth
RUN ["/bin/mkdir", "$SRV_DIR"]
WORKDIR $SRV_DIR 
COPY ./spotify_auth $SRV_DIR 
COPY ./server.crt $SRV_DIR
COPY ./server.key $SRV_DIR
# USER nobody
ENTRYPOINT ["./spotify_auth"]
CMD ["9000"]

LABEL 'spotify_auth_port'='9000'


