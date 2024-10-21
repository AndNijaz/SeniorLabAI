FROM python:3.10

WORKDIR /app

COPY requirements.txt .
RUN pip install -r requirements.txt

COPY . .

ENV FLASK_APP=app.py
ENV FLASK_RUN_PORT=8468
EXPOSE 8468

CMD ["flask", "run", "--host=0.0.0.0"]
