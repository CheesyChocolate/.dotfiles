# Change the user permission of files accordingly

configs:	chmod 644
scripts:	chmod 755

# Install the packages

packages explicitly required by the system are listed in official_pkg_list.txt

```bash
pacman -S --needed - < official_pkg_list.txt
```

optional packages are listed in optional_pkg_list.txt

```bash
pacman -S --needed --asdeps - < optional_pkg_list.txt
```

packages from the AUR and other foreign sources are listed in foreign_pkg_list.txt

```bash
yay -S --needed - < foreign_pkg_list.txt
```
