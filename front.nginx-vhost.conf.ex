
server {
    server_name backup.example.com;

    root /usr/share/bakapy/front;

    auth_basic "Authentication required";
    auth_basic_user_file /etc/bakapy/front.pw;

    # Files storage
    location /storage {
        alias /var/lib/bakapy/storage;
    }

    # Metadata storage. Autoindex required for webui.
    location /metadata {
        autoindex on;
        alias /var/lib/bakapy/metadata;
    }
}
