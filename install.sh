#!/bin/bash

rc_local_file="/etc/rc.local"
actual_user=$(logname)
user_home=$(eval echo ~${actual_user})
eis_command="cd $user_home/eis && nohup ./eis &"

if [ -f "$rc_local_file" ]; then
    if ! grep -q "$eis_command" "$rc_local_file"; then
        sed -i "/^exit 0/i $eis_command" "$rc_local_file"
        echo "Added Done"
    else
        echo "EIS command already exists in $rc_local_file"
    fi
else
    echo -e "#!/bin/sh -e\n$eis_command\nexit 0" > "$rc_local_file"
    chmod +x "$rc_local_file"
    echo "Created $rc_local_file with EIS command"
fi
