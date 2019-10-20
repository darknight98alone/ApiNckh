# ApiNckh  
- API để lưu file cho phép đẩy file lên, trả về text và ID, thêm mức độ để xử lý file đẩy lên, (deskew, deblur, autofill, xử lý bảng,(lưu file và text cùng thư mục temp)
- [x] API nhận text và ID, đẩy lên elastic search.
api có dạng: localhost:8080//pushtextandid  
json client gửi lên có định dạng:  
  {  
  &nbsp;&nbsp;&nbsp;"mac":"123456sodp",  
  &nbsp;&nbsp;&nbsp;"id":"1",  
  &nbsp;&nbsp;&nbsp;"contents":"noi dung cua file"  
  }  
- [x] API gửi lên đoạn text, trả về danh sách các tên file, nội dung của file được giới hạn 100 từ và ID đi kèm
api có dạng: localhost:8080//search    
json client gửi lên có dạng:  
{  
	&nbsp;&nbsp;&nbsp;"mac":"documents",  
	&nbsp;&nbsp;&nbsp;"search_contents":"khong co viec gi kho"  
}  
json trả về có dạng mảng:  
[
    {  
        &nbsp;&nbsp;&nbsp;"_source": {  
            &nbsp;&nbsp;&nbsp;"id": "1",  
            &nbsp;&nbsp;&nbsp;"contents": "khong co",  
            &nbsp;&nbsp;&nbsp;"filename": "test.txt"  
        &nbsp;&nbsp;&nbsp;}  
    },  
    {  
        &nbsp;&nbsp;&nbsp;"_source": {  
            &nbsp;&nbsp;&nbsp;"id": "2",  
            &nbsp;&nbsp;&nbsp;"contents": "không có việc gì khó chỉ sợ lòng không bền",  
            &nbsp;&nbsp;&nbsp;"filename": "temp.txt"  
        &nbsp;&nbsp;&nbsp;}  
    }
]  
- API nhận ID, chỉ trả về file text
api: localhost:8080/getAllContents  
json client tải lên có dạng:    
{  
	"id":"1"  
}  
- API Download nhận id từ client và trả file gốc về:  
> api have format: localhost:8080/download  
> json client send have format:  
>{  
"mac": "1234",  
"id": "123",  
}  
+ trước tiên phải gọi tới Api con trả về extension của file gốc
api có dạng: localhost:8080/getRootFileExtension  
json client tải lên có dạng:    
{  
	"id":"1"  
}
+ sau đó sẽ gọi tới Api download file gốc về  
api có dạng: localhost:8080/download  
json client tải lên có dạng:    
{  
	"id":"1"  
}  
