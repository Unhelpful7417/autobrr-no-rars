#!/bin/ash

# Get torrent file path from autobrr's {{.TorrentPathName}} variable
torrentFile=$1

# Get .rar file tolerance from user
rarTolerance=$2

# Use imdl to get a JSON dump of the torrent's metadata
torrentJSON=$(/app/imdl torrent show -j -i $torrentFile)

# Check if any file paths end with ".rar"
result=$(echo "$torrentJSON" | jq -r '.files | values[] | select(endswith(".rar"))')

# If the number of rars found is not within user-defined tolerance, then
# exit with non-standard code to signify a rar'd release was grabbed
if [ -n "$result" ]; then
  # Count the number of lines
  rarCount=$(echo "$result" | wc -l)

  if [ "$rarCount" -gt "$rarTolerance" ]; then
    exit 3 # Fail since there were more rar files than the number allowed
  else
    exit 0 # Pass since the number of rar files was within the tolerance
  fi
else
  exit 0 # Pass since there were no rar files found at all
fi

