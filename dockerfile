FROM python:latest
LABEL Maintainer="Timendus"
COPY . .
RUN pip3 install -r requirements.txt
CMD [ "python3", "-m", "meshbot" ]
