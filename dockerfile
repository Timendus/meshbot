FROM python:latest
LABEL Maintainer="Timendus"
COPY . .
RUN pip install python-dotenv pytap2 meshtastic
CMD [ "python3", "./main.py" ]
