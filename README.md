# TextOverImg
Naive implementation of an app that users can submit text and an image URL to.
The app returns the image with the text placed over it.


Example queries to check endpoints:

curl -X POST -d "{\"url\": \"https://upload.wikimedia.org/wikipedia/commons/3/3d/Forstarbeiten_in_%C3%96sterreich.JPG\", \"text\": \"Inpsiration Quote Here!\"}" http://localhost:3000/image

curl -X POST -d "{\"username\": \"test\", \"password\": \"test\"}" http://localhost:3000/user

TODO:	
- User db and logout functions
- Frontend in Vue.js
- Improve responses from API
    -> display errors to the user on the frontend, not just in the console as its currently doing
        -> e.g. "No image found on that url", "Incorrect login", etc...
