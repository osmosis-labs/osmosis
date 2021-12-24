<template>
    <div class="dark-mode-widget">
        <!-- <input type="checkbox" id="theme-toggle" @click="toggleDarkTheme"></input>
        <label for="theme-toggle"><span></span></label> -->
    </div>
</template>

<script>
const options = {
  bottom: '64px', // default: '32px'
  right: 'unset', // default: '32px'
  left: '32px', // default: 'unset'
  time: '0.5s', // default: '0.3s'
  mixColor: '#fff', // default: '#fff'
  backgroundColor: '#fff',  // default: '#fff'
  buttonColorDark: '#FFF',  // default: '#100f2c'
  buttonColorLight: '#fff', // default: '#fff'
  saveInCookies: false, // default: true,
  label: '🌓', // default: ''
  autoMatchOsTheme: true // default: true
}

import Darkmode from 'darkmode-js';
const darkmode = new Darkmode(options);
darkmode.showWidget();

// new Darkmode().showWidget();

export default {
name: 'ToggleDarkMode',
mounted() {
    this.checkUserPreference();
},
methods: {
    toggleDarkTheme() {
    const body = document.body;
//     body.classList.toggle("dark-mode");
//     //If dark mode is selected
//     if (body.classList.contains("dark-mode")) {
//         //Save user preference in storage
//         localStorage.setItem("dark-theme", "true");
//     //If light mode is selected
//     } else {
//         body.classList.remove("dark-mode");
//         setTimeout(function() {
//         localStorage.removeItem("dark-theme");
//         }, 100);
//     }
    },
    checkUserPreference(){
        //Check Storage on Page load. Keep user preference through sessions
    //     if (localStorage.getItem("dark-theme")) {
    //         document.body.classList.add("dark-mode");
    //         document.getElementById('theme-toggle').checked = true;
    //     }
    }
}
};
</script>
