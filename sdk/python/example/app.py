# app.py
# Example Flask application importing and using the Limiter.io Python SDK

import os
import sys
from flask import Flask, jsonify

# Add parent directory to sys.path to import local SDK modules
sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from client import LimiterClient
from decorator import flask_rate_limit

app = Flask(__name__)

# 1. Initialize the official SDK client
client = LimiterClient('http://localhost:8080', 'replace_with_developer_api_key')

# 2. Mount endpoints and apply Flask SDK decorator
@app.route('/api/payments', methods=['GET'])
@flask_rate_limit(client)
def get_payments():
    return jsonify({
        "status": "success",
        "message": "Payment processed successfully using Python SDK!"
    })

@app.route('/api/data', methods=['GET'])
@flask_rate_limit(client)
def get_data():
    return jsonify({
        "status": "success",
        "message": "Telemetry metrics loaded using Python SDK."
    })

if __name__ == '__main__':
    port = int(os.environ.get('PORT', 9002))
    app.run(port=port)
