#!/usr/bin/python3

import os
import sys
import requests
from pathlib import Path
import tkinter as tk
from tkinter import filedialog, messagebox, simpledialog

def select_files():
    """Open a file dialog to select JPG/PNG files"""
    root = tk.Tk()
    root.withdraw()  # Hide the main window

    file_types = [
        ("Image files", "*.jpg *.jpeg *.png *.JPG *.JPEG *.PNG"),
        ("JPEG files", "*.jpg *.jpeg *.JPG *.JPEG"),
        ("PNG files", "*.png *.PNG"),
        ("All files", "*.*")
    ]

    files = filedialog.askopenfilenames(
        title="Select images to upload",
        filetypes=file_types
    )

    root.destroy()
    return files

def get_upload_params():
    """Get upload parameters from user"""
    root = tk.Tk()
    root.withdraw()

    # Get max dimension
    max_dimension = simpledialog.askinteger(
        "Upload Settings",
        "Enter maximum dimension (pixels):",
        initialvalue=1920,
        minvalue=100,
        maxvalue=4000
    )

    if max_dimension is None:
        max_dimension = 1920

    # Get folder name
    folder = simpledialog.askstring(
        "Upload Settings",
        "Enter folder name (optional):",
        initialvalue=""
    )

    if folder is None:
        folder = ""

    root.destroy()
    return max_dimension, folder

def upload_images(files, max_dimension=1920, folder=""):
    """Upload selected files to the API endpoint"""
    if not files:
        print("No files selected.")
        return False

    print(f"Uploading {len(files)} files...")

    # Prepare form data
    form_data = {
        'maxDimension': str(max_dimension),
        'folder': folder
    }

    files_data = []
    for file_path in files:
        try:
            with open(file_path, 'rb') as f:
                files_data.append(('images', (os.path.basename(file_path), f.read(), 'image/jpeg')))
        except Exception as e:
            print(f"Error reading file {file_path}: {e}")
            return False

    try:
        response = requests.post(
            "https://davispics.com/upload/",
            data=form_data,
            files=files_data,
            timeout=300  # 5 minute timeout
        )

        if response.ok:
            print("Images uploaded successfully!")
            return True
        else:
            try:
                error_data = response.json()
                error_msg = error_data.get('message', 'Upload failed')
            except:
                error_msg = f"Upload failed with status {response.status_code}"
            print(f"Error: {error_msg}")
            return False
    except requests.exceptions.RequestException as e:
        print(f"Network error occurred: {e}")
        return False

def main():
    """Main function to run the upload script"""
    print("Davis Pics Upload Script")
    print("========================")

    # Check if tkinter is available
    try:
        import tkinter as tk
    except ImportError:
        print("Error: tkinter is required but not available.")
        print("Please install tkinter: sudo apt-get install python3-tk")
        sys.exit(1)

    # Select files
    print("Opening file dialog...")
    selected_files = select_files()

    if not selected_files:
        print("No files selected. Exiting.")
        return

    print(f"Selected {len(selected_files)} files:")
    for file_path in selected_files:
        print(f"  - {os.path.basename(file_path)}")

    # Get upload parameters
    max_dimension, folder = get_upload_params()

    print(f"\nUpload settings:")
    print(f"  Max dimension: {max_dimension}px")
    print(f"  Folder: '{folder}' {'(root)' if not folder else ''}")

    # Confirm upload
    confirm = input("\nProceed with upload? (y/N): ").strip().lower()
    if confirm not in ['y', 'yes']:
        print("Upload cancelled.")
        return

    # Upload files
    success = upload_images(selected_files, max_dimension, folder)

    if success:
        print("\n✓ Upload completed successfully!")
    else:
        print("\n✗ Upload failed!")
        sys.exit(1)

if __name__ == "__main__":
    main()
