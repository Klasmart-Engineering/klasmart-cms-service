FROM alpine
COPY ["jwt_public_key.pem", "ams_assessment_jwt_public_key.pem", "h5p_public_key.pem", "live_token_private_key.pem", "h5p_private_key.pem", "./"]
COPY ./handler /handler
ENTRYPOINT [ "/handler" ]
