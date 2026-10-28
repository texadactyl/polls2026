set -e

rm -f plots/*
rm -f logs/*

polls2026 -l
polls2026 -r 2 0
polls2026 -r 2 2
polls2026 -r 2 4
polls2026 -p ak
polls2026 -p fl
polls2026 -p la
polls2026 -p ia
polls2026 -p ks
polls2026 -p me
polls2026 -p mi
polls2026 -p mn
polls2026 -p ms
polls2026 -p mt
