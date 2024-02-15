import os

def main():
    print("(enter q to cancel at any time)")
    folder = input("Folder path: ")
    if folder.lower() == "q":
        print("cancelled")
        return
    basename = input("file rename base: ")
    if basename.lower() == "q":
        print("cancelled")
        return
    ext = input("file extension (w/o period): ")
    if ext.lower() == "q":
        print("cancelled")
        return
    print(f"folder: {folder}")
    print(f"new name scheme: {basename}1.{ext}")
    cont = input("continue? [y or n]: ")
    if cont.lower() != "y":
        print("cancelled")
        return
    count = 0
    for _, filename in enumerate(os.listdir(folder)):
        src = os.path.join(folder, filename)
        if not src.endswith(ext):
            print(f"error: original file does not end with {ext} extension!")
            print(src)
            print("skipping")
            continue
        newName = f"{basename}{count}.{ext}"
        dst = os.path.join(folder, newName)
        count += 1
        os.rename(src, dst)

if __name__ == '__main__':
    main()