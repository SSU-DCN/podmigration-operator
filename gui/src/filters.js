import Vue from "vue";

Vue.filter("capitalize", function(value) {
  if (!value) return "";
  return (
    value
      .toString()
      .charAt(0)
      .toUpperCase() + value.slice(1)
  );
});

Vue.filter("titleCase", function(value) {
  if (!value) return "";
  return value.toString().replace(/([^\W_]+[^\s-]*) */g, function(s) {
    return s.charAt(0).toUpperCase() + s.substr(1).toLowerCase();
  });
});
