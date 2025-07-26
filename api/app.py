from flask import Flask, jsonify
from flask_cors import CORS
import os

app = Flask(__name__)
CORS(app)

# Configuration from environment variables
app.config['COUCHDB_URL'] = os.getenv('COUCHDB_URL', 'http://localhost:5984')
app.config['IPFS_API_URL'] = os.getenv('IPFS_API_URL', 'http://localhost:5001')

@app.route('/')
def health_check():
    return jsonify({
        "status": "LedgerLynx API is running",
        "couchdb_url": app.config['COUCHDB_URL'],
        "ipfs_url": app.config['IPFS_API_URL']
    })

@app.route('/api/status')
def api_status():
    return jsonify({
        "api": "online",
        "services": {
            "couchdb": "connected",
            "ipfs": "connected"
        }
    })

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000, debug=True)
