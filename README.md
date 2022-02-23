# TextOverImg
Naive implementation of an app that users can submit text and an image URL to.
The app returns the image with the text placed over it.


User request example:
curl -X POST -d "{\"url\": \"https://upload.wikimedia.org/wikipedia/commons/3/3d/Forstarbeiten_in_%C3%96sterreich.JPG\", \"text\": \"Inpsiration Quote Here!\"}" http://localhost:3000/req

User login example:
curl -X POST -d "{\"username\": \"test\", \"password\": \"test\"}" http://localhost:3000/user

TODO:	
- User db and login/logout functions
- Frontend in Vue.js
