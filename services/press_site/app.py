import datetime
import os
import urlparse

from flask import Flask, render_template, request, redirect
from flask.ext.sqlalchemy import SQLAlchemy

app = Flask(__name__)

app.config['SQLALCHEMY_DATABASE_URI'] = os.getenv('DATABASE_URL', 'sqlite:///app.sqlite3')
db = SQLAlchemy(app)

class Email(db.Model):
	id = db.Column(db.Integer, primary_key=True, autoincrement=True)
	email = db.Column(db.String, nullable=False)
	email_added_on = db.Column(db.DateTime, nullable=False)

	def __init__(self, email):
		self.email = email
		self.email_added_on = datetime.datetime.now()

@app.route('/')
def index():
	return render_template('index.html')

@app.route('/signup', methods=['POST'])
def signup():
	email = Email(request.form['email'])
	db.session.add(email)
	db.session.commit()

	return redirect('/')

@app.errorhandler(404)
def not_found(exception):
	return render_template('404.html'), 404

if __name__ == '__main__':
    app.run()
