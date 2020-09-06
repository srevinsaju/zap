import flask
from flask import request, Request
import requests
from flask import abort
from flask import jsonify
import json
from zap.appimage.generator import AppImageConfigJsonGenerator
app = flask.Flask(__name__)
app.config['DEBUG'] = True

FEED_URL = "http://appimage.github.io/feed.json"

def get_feed_json():
    feed_request = requests.get(FEED_URL)
    return feed_request.json()

def get_app_data_from_feed(appname):
    feed = get_feed_json().get('items')
    for appdata in feed:
        if appdata.get('name').lower() == appname.lower():
            return appdata
        elif appdata.get('name').replace('_', '-') == appname.lower().replace('_', '-'):
            return appdata


@app.route('/core/<appname>', methods=['GET'])
def return_data(appname):
    appdata = get_app_data_from_feed(appname)
    if appdata is None:
        abort(404)
    app_core_json = AppImageConfigJsonGenerator(appdata)
    return jsonify(
            json.loads(
                json.dumps(
                    app_core_json.get_app_metadata())))


    
if __name__ == "__main__":
    app.run()
