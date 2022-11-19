# change the user permision of files accrodinglly

configs:	chmod 644
scripts:	chmod 755

# install packages from a previously saved list of packages, while not reinstalling previously installed packages that are already up-to-date

pacman -S --needed - < offical_pkg_list.txt
pacman -S --needed --asdeps - < optional_pkg_list.txt
yay -S --needed - < foreign_pkg_list.txt
