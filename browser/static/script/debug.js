export const local = ({
    "": true, "localhost": true, "127.0.0.1": true, '[::1]': true,
})[document.location.hostname];
