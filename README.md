# ApiNckh  
- API để lưu file cho phép đẩy file lên, trả về text và ID, thêm mức độ để xử lý file đẩy lên, (deskew, deblur, autofill, xử lý bảng,
- [x] API nhận text và ID, đẩy lên elastic search.
json client gửi lên có định dạng:  
  {  
  "mac":"123456sodp",  
  "id":"1",  
  "contents":"noi dung cua file"  
  }  
- [x] API gửi lên đoạn text, trả về danh sách các tên file, nội dung của file được giới hạn 100 từ và ID đi kèm
json client gửi lên có dạng:  
{  
	"mac":"documents",  
	"search_contents":"khong co viec gi kho"  
}  
json trả về có dạng mảng:  
[  
    {  
        "_source": {  
            "id": "1",  
            "contents": "khong co",  
            "filename": "test.txt"  
        }  
    },  
    {  
        "_source": {  
            "id": "2",  
            "contents": "không có việc gì khó chỉ sợ lòng không bền",  
            "filename": "temp.txt"  
        }  
    }  
]  
- API nhận tên file và ID, chỉ trả về file text
- API Download sẽ tải file gốc về
