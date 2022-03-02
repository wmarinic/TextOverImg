<template>
  <div id="app" class="container">
    <div class="column left">
        <h1>Create an Inspirational Image</h1>
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
        <p id="img_err"></p>
        <img :src="img"/> 
    </div>  
    <div class="column right">
      <div id="login">
        <h3>Login</h3>
        <form v-on:submit.prevent="userLogin">
          <div class="form-group">
            <input v-model="user" type="text" id="username-input" placeholder="Username" class="form-control">
            <br>
            <input v-model="pass" type="text" id="password-input" placeholder="Password" class="form-control">
          </div>
          <div class="form-group">
            <button class="btn btn-primary">Login</button>
          </div>
        </form>
        <br>
        <a href="register.html">Click here to create an account</a>
        <p id="login_msg"></p>
      </div>
      <div id="logged_in" style="display:none">
        <p id="display_username"></p>
        <button v-on:click="userLogout">Logout</button>
      </div>
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
    img: '',
    user: '',
    pass: '',
    premium: false,
    wait: false
  } },

  methods: {
    makeInspirationalImg() {
      //prevent users from quickly sending multiple requests
      if (this.wait){
        return;
      }
      this.wait = true;
      setTimeout(() => this.wait = false, 1000)
      //reset the image and img_error
      this.img = '';
      document.getElementById("img_err").innerHTML = "";

      //post to the go api
      axios.post("http://localhost:3000/image", {
        url: this.imageURL,
        text: this.text,
        auth: this.premium,
      })
      .then((response) => {
        //check for error
        if(response.data.error == "none"){
          this.img = response.data.image;
        } else{
          document.getElementById("img_err").innerHTML = response.data.error;
        }
      })
      .catch((error) => {
        window.alert(`API error: ${error}`);
      })
    },
    userLogin(){
      //reset err messages
      document.getElementById("login_msg").innerHTML = "";
      //post to the go api
      axios.post("http://localhost:3000/login", {
        username: this.user,
        password: this.pass,
      })
      .then((response) =>{
        if(response.data.status == "success"){
          this.premium = true;
          //hide the login form and display a logout button
          document.getElementById("login").style.display = "none";
          document.getElementById("display_username").innerHTML = "Logged in as: " + response.data.user;
          document.getElementById("logged_in").style.display = "initial";
        }else{
          document.getElementById("login_msg").innerHTML = response.data.msg;
        }
      })
      .catch((error) => {
        window.alert(`Login API error: ${error}`);
      })
    },
    userLogout(){
      this.premium = false;
      document.getElementById("login").style.display = "initial";
      document.getElementById("logged_in").style.display = "none";
    }
  },
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
