<template>
  <div id="app" class="container">
    <div class="column left">
        <h1>Create an inspirational image</h1>
        <form v-on:submit.prevent="makeInspirationalImg">
          <div class="form-group">
            <input v-model="imageURL" type="text" id="url-input" placeholder="Enter an image URL" class="form-control">
            <br>
            <input v-model="text" type="text" id="text-input" placeholder="Enter text" class="form-control">
          </div>
          <div class="form-group">
            <button class="btn btn-primary">Create Inspirational Image!</button>
          </div>
        </form>
        <img :src="image"/> 
    </div>  
    <div class="column right">
      <h3>Login to Your Account</h3>
      <form v-on:submit.prevent="userLogin">
        <div class="form-group">
          <input v-model="user" type="text" id="username-input" placeholder="Enter username" class="form-control">
          <br>
          <input v-model="pass" type="text" id="password-input" placeholder="Enter password" class="form-control">
        </div>
        <div class="form-group">
          <button class="btn btn-primary">Login</button>
        </div>
      </form>

    </div>
  </div>
</template>

<script>
import axios from 'axios';

export default {
  name: 'App',

  data() { return {
    imageURL: '',
    text: '',
    image: '',
    user: '',
    pass: '',
  } },

  methods: {
    makeInspirationalImg() {
      //reset the image
      this.image = '';
      //post to the go api
      axios.post("http://localhost:3000/image", {
        url: this.imageURL,
        text: this.text,
      })
      .then((response) => {
         this.image = response.data;
      })
      .catch((error) => {
        window.alert(`API error: ${error}`);
      })
    },
    userLogin(){
      //post to the go api
      axios.post("http://localhost:3000/user", {
        username: this.user,
        password: this.pass,
      }) //TODO add response from login
      .catch((error) => {
        window.alert(`Login API error: ${error}`);
      })
    }
  }  
}
</script>

<style>
#app {
  font-family: Avenir, Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  text-align: center;
  color: #2c3e50;
  margin-top: 60px;
}

.column {
  float:left;
}
.left {
  width: 80%;
}

.right {
  width: 20%;
}

@media screen and (max-width: 600px){
  .column{
    width:100%;
  }
};

</style>
