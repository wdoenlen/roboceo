import asana
import constants

def build_asana_client():
    """Connects to Asana and returns an API client"""
    return asana.Client.access_token(constants.CEO_ACCESS_TOKEN)
