lobster_editor=nvim
player=mpv
download_dir=$HOME/Videos/Lobster
provider=Vidcloud
history=1
subs_language=english

quality="1080"

json_from_id() {
	json_data=$(curl -s "http://127.0.0.1:8888/.netlify/functions/decrypt?id=${source_id}")
}
