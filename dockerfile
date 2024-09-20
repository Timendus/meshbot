FROM python:latest
LABEL Maintainer="Timendus"
COPY . .
RUN pip install pytap2 meshtastic
CMD [ "python3", "./main.py", "meshtastic.local" ]
