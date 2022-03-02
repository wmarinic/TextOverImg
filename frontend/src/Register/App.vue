<template>
  <div id="app" class="container">
      <h3>Create New Account</h3>
      <form v-on:submit.prevent="userRegister">
        <div class="form-group">
          <input v-model="user" type="text" id="username-input" placeholder="Username" class="form-control">
          <br>
          <input v-model="pass" type="text" id="password-input" placeholder="Password" class="form-control">
        </div>
        <div class="form-group">
          <button class="btn btn-primary">Sign Up</button>
        </div>
      </form>
      <p id="register_msg"></p>
  </div>
</template>

<script>
import axios from 'axios';

export default {
  name: 'App',

  data() { return {
    user: '',
    pass: '',
  } },

  methods: {
    userRegister(){
      //reset err messages
      document.getElementById("register_msg").innerHTML = "";
      //post to the go api
      axios.post("http://localhost:3000/register", {
        username: this.user,
        password: this.pass,
      })
      .then((response) =>{
        //display response
        document.getElementById("register_msg").innerHTML = response.data.msg;
      })
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
</style>