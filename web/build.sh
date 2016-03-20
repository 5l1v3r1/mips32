cd src
rm -f src.js
rm -f src.js.map
gopherjs build
mv src.js ../assets/src.js
mv src.js.map ../assets/src.js.map
