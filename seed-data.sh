#!/bin/bash

# TrailMemo API - Seed Script
# Seeds the database with realistic test data

set -e  # Exit on error

# Configuration
API_URL="${API_URL:-http://localhost:8080}"
TOKEN="${TOKEN}"

if [ -z "$TOKEN" ]; then
  echo "❌ Error: TOKEN environment variable not set"
  echo "Run this first:"
  echo "  export TOKEN=\"your_firebase_token\""
  exit 1
fi

echo "🌱 Seeding TrailMemo database..."
echo "📍 API: $API_URL"
echo ""

# Array of park names
PARKS=(
  "Lindley Park"
  "Peets Hill"
  "Gallagator Trail"
  "Burke Park"
  "Bozeman Creek Trail"
  "Drinking Horse Mountain"
  "Main Street to the Mountains"
  "Sourdough Trail"
  "Triple Tree Trail"
  "Story Mill Park"
)

# Array of memo texts (realistic trail observations)
TEXTS=(
  "Found fallen tree blocking main trail near north entrance. Approximately 2 feet in diameter. Will need chainsaw to clear."
  "Trail marker at junction is damaged and pointing wrong direction. Need to replace or reorient."
  "Graffiti on picnic table at viewpoint. Black spray paint. Table also needs restaining."
  "Erosion on steep section creating water runoff issues. Recommend adding water bars or rerouting."
  "Mountain bike jump feature appears unsafe. Logs are rotting. Should be removed or rebuilt."
  "Invasive knapweed spreading along trail corridor. Needs treatment before it spreads further."
  "Bench at mile 2 has broken slats. One person already injured. Immediate repair needed."
  "Trash overflowing at trailhead. Also noticed bear activity near dumpster. Consider bear-proof container."
  "New social trail forming parallel to main trail. Causing erosion and vegetation damage. Need barriers."
  "Trail signage faded and illegible. Hikers getting confused at junction. Replace signs."
  "Wasp nest forming under bridge structure. Right at head height. Safety hazard for trail users."
  "Amazing wildflower bloom in meadow area. Would be great spot for interpretive sign about native plants."
  "Dog waste bags dispenser empty. Seeing increase in waste left on trail. Need refill."
  "Parking lot pothole getting worse. Could damage vehicles. Need asphalt repair."
  "Stunning wildlife sighting - moose with calf near creek. Great spot for wildlife viewing area."
  "Trail bridge has loose planks. Heard rattling when crossing. Safety inspection needed."
  "Excellent volunteer turnout for trail work day. Cleared 2 miles of overgrowth. Great progress!"
  "Muddy section near creek is widening as people walk around it. Consider boardwalk or raised tread."
  "Poison ivy growing along trail edge. Need to mark area and add warning sign."
  "Beautiful new trail section completed. Drainage is working well and tread is solid."
  "Root exposure on popular section creating trip hazards. Need soil addition and trail hardening."
  "Directional arrow carved into tree. Permanent scarring. Document and remove."
  "Great family feedback on new nature scavenger hunt signs. Kids really engaged."
  "Rockfall on upper trail section. Boulders on trail. Need equipment to move."
  "Excellent condition after maintenance crew work. Trail is in best shape in years."
  "Mountain lion scat and tracks observed. Fresh. Alert users about increased activity."
  "New unauthorized camping spot with fire ring. In sensitive riparian area. Need restoration."
  "Wheelchair accessible portion of trail working great. Received positive feedback from users."
  "Invasive species removal project successful. Native plants rebounding nicely."
  "Trail counter showing 50% increase in usage this year. Consider capacity management."
)

# Bozeman area coordinates (varying locations)
LOCATIONS=(
  "45.6769:-111.0429"  # Lindley Park area
  "45.6895:-111.0234"  # Peets Hill area
  "45.6701:-111.0651"  # Gallagator area
  "45.6834:-111.0423"  # Burke Park area
  "45.6612:-111.0389"  # Creek area
  "45.7123:-111.0567"  # Drinking Horse area
  "45.6789:-111.0389"  # Downtown trails area
  "45.7234:-111.0712"  # Sourdough area
  "45.6543:-111.0823"  # Triple Tree area
  "45.6912:-111.0334"  # Story Mill area
)

# Create memos
COUNT=0
TOTAL=30

echo "Creating $TOTAL memos..."

for i in $(seq 1 $TOTAL); do
  # Random selections
  PARK="${PARKS[$((RANDOM % ${#PARKS[@]}))]}"
  TEXT="${TEXTS[$((RANDOM % ${#TEXTS[@]}))]}"
  LOCATION="${LOCATIONS[$((RANDOM % ${#LOCATIONS[@]}))]}"
  
  # Parse location
  LAT=$(echo $LOCATION | cut -d: -f1)
  LON=$(echo $LOCATION | cut -d: -f2)
  
  # Random duration (10-120 seconds)
  DURATION=$((10 + RANDOM % 110))
  
  # Random accuracy (5-20 meters)
  ACCURACY=$((5 + RANDOM % 15))
  
  # Determine if it needs urgent attention (random)
  URGENCY=$((RANDOM % 100))
  if [ $URGENCY -lt 20 ]; then
    TITLE="URGENT: ${TEXT:0:40}..."
  else
    TITLE="${TEXT:0:50}..."
  fi
  
  # Create memo
  RESPONSE=$(curl -s -X POST $API_URL/api/v1/memos \
    -H "Authorization: Bearer $TOKEN" \
    -F "text=$TEXT" \
    -F "duration_seconds=$DURATION" \
    -F "latitude=$LAT" \
    -F "longitude=$LON" \
    -F "location_accuracy=$ACCURACY" \
    -F "park_name=$PARK" \
    -F "title=$TITLE")
  
  # Check if successful
  if echo "$RESPONSE" | grep -q "memo_id"; then
    COUNT=$((COUNT + 1))
    echo "✅ Created memo $COUNT/$TOTAL: $PARK - ${TITLE:0:30}..."
  else
    echo "❌ Failed to create memo: $RESPONSE"
  fi
  
  # Small delay to avoid rate limiting
  sleep 0.2
done

echo ""
echo "🎉 Seeding complete!"
echo "📊 Created $COUNT/$TOTAL memos"
echo ""
echo "Test your API:"
echo "  curl $API_URL/api/v1/memos -H \"Authorization: Bearer \$TOKEN\""
echo ""
echo "View on map:"
echo "  Open your iOS app and see all the pins!"

